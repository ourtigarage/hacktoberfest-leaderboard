require 'concurrent'
require 'faraday-http-cache'
require 'json'
require 'octokit'
require 'open-uri'

include Concurrent

ORG_URL = 'https://api.github.com/repos/ourtigarage'.freeze
SNAKE_URL = "#{ORG_URL}/web-snake".freeze
LEADERBOARD_URL = "#{ORG_URL}/hacktoberfest-leaderboard".freeze

# Class reprensenting a badge a player can earn
class Badge
  attr_reader :short, :title, :description

  def initialize(short, title, description, &block)
    @short = short
    @title = title
    @description = description
    @block = block
  end

  def earned_by?(player)
    @block.call(player)
  end
end

BADGES = [
  Badge.new('medal', 'Completed hacktoberfest', 'The player completed the hacktoberfest challenge by submitting 4 pull requests', &:challenge_complete?),
  Badge.new('snake', 'Snake charmer', 'The player submitted at least 1 PR to the snake game', &:contributed_to_snake?),
  Badge.new('leaderboard', 'Leaderboard contributor', 'The player submitted at least 1 PR to the leaderboard code', &:contributed_to_leaderboard?),
  Badge.new('10-contributions', 'Pull Request champion', 'The player submitted more than 10 Pull requests', &:ten_contributions?),
  Badge.new('adventure', 'Adventurer', 'The player submitted at least 1 PR to a repository out of "ourtigarage" organisation', &:contributed_out_of_org?)
].freeze

# The leaderboard root class, where the magic happens
class Leaderboard
  # Initialize the leaderboard for the given event date and participant file URL
  def initialize(event_date, participants_file_url)
    @file_uri = URI(participants_file_url)
    @event_date = event_date
    # Connect to GitHub using a token from env variable.
    # If no token is set, no problem it will still work,
    # but with a limited rate for API calls
    @github = Octokit::Client.new access_token: ENV['GH_TOKEN']
    @github.auto_paginate = true # Use to concatenate all results
    @github.middleware = Faraday::RackBuilder.new do |builder|
      builder.use Faraday::HttpCache, serializer: Marshal, shared_cache: false
      builder.use Octokit::Response::RaiseError
      builder.adapter Faraday.default_adapter
    end
  end

  # Retrieve the list of participants names from GitHub page
  def members_names
    # Extract usernames from file
    open(@file_uri).each_line
                   .lazy
                   .map(&:strip)
                   .map { |l| l.match(/^\* .*@([a-zA-Z0-9]+).*$/) }
                   .reject(&:nil?)
                   .map { |m| m[1] }
                   .reject { |name| name == 'username' }
                   .to_a
                   .uniq
  end

  # Build a list of members with additional data from GitHub
  # Those data are user info like name, link to avatar and profiles
  # and the list of contributions during the hacktoberfest's month
  def members
    members_data(members_names).values
                               .map { |user| Member.new(*user) }
  end

  private

  def query_users_data(usernames)
    authors = usernames.map { |n| "author:#{n}" }.join ' '
    query_filter = "is:pr #{authors} created:#{@event_date} -label:invalid"
    @github.search_issues(query_filter)
           .items
           .each_with_object({}) { |e, acc| (acc[e.user.login] ||= [e.user, []])[1] << e }
  end

  def query_missing_users_data(usernames, data)
    usernames.reject { |name| data.include? name }
             .map { |u| Future.execute { get_member_from_github u } }
             .each { |f| raise f.reason if !f.value && f.rejected? }
             .map(&:value)
             .reject(&:nil?)
  end

  def add_missing_users(usernames, data)
    query_missing_users_data(usernames, data)
      .each_with_object(data) { |e, acc| add_user(acc, e) }
  end

  def members_data(usernames)
    add_missing_users(usernames, query_users_data(usernames))
  end

  def get_member_from_github(username)
    @github.user(username)
  rescue Octokit::NotFound
    nil
  end
end

# A contest member from the participant list in landing page
class Member
  attr_reader :username, :avatar, :profile, :contributions

  # Construct a user using the data fetched from GitHub
  def initialize(github_user, contributions)
    @username = github_user.login
    @avatar = github_user.avatar_url
    @profile = github_user.html_url
    @contributions = contributions
  end

  # Check if the user has completed the challenge
  def challenge_complete?
    contributions.size >= 4
  end

  # Returns the completion percentage
  def challenge_completion
    [100, ((contributions_count.to_f / 4.0) * 100).to_i].min
  end

  # Count the number of valid contributions
  def contributions_count
    contributions.size
  end

  def contributed_to_snake?
    contributions.any? { |c| c.repository_url == SNAKE_URL }
  end

  def contributed_to_leaderboard?
    contributions.any? { |c| c.repository_url == LEADERBOARD_URL }
  end

  def ten_contributions?
    contributions.size >= 10
  end

  def contributed_out_of_org?
    contributions.any? { |c| !c.repository_url.start_with? ORG_URL }
  end

  def badges
    BADGES.select { |b| b.earned_by?(self) }
  end

  def to_json(*_opts)
    {
      username: @username,
      avatar: @avatar,
      profile: @profile
    }.to_json
  end
end
