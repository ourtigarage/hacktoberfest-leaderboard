require 'concurrent'
require 'faraday-http-cache'
require 'json'
require 'octokit'
require 'open-uri'
require_relative 'badge'
require_relative 'member'

include Concurrent

BASE_API_URL = 'https://api.github.com'.freeze
BASE_REPOS_URL = "#{BASE_API_URL}/repos".freeze
ORG_REPOS_URL = "#{BASE_REPOS_URL}/ourtigarage".freeze
SNAKE_URL = "#{ORG_REPOS_URL}/web-snake".freeze
LEADERBOARD_URL = "#{ORG_REPOS_URL}/hacktoberfest-leaderboard".freeze

# This is the list of all badges that can be earned
# The block that define the badge's challenge should return either an integer or a boolean
BADGES = [
  Badge.new('hacktoberfest', 'Hacktoberfest completed', 'Completed the hacktoberfest challenge by submitting 4 pull requests', &:challenge_complete?),
  Badge.new('snake', 'The snake charmer', 'Submitted 1 Pull Request to the <a href="https://ourtigarage.github.io/web-snake/">snake game</a>\'s code repository', &:contributed_to_snake),
  Badge.new('leaderboard', 'The leaderboard contributor', 'Submitted 1 Pull Request to this leaderboard\'s code repository', &:contributed_to_leaderboard),
  Badge.new('10-contributions', 'The Pull Request champion', 'Submitted more than 10 Pull requests', &:ten_contributions?),
  Badge.new('adventure', 'The adventurer', 'Submitted 1 Pull Request to a repository he does not own, out of <a href="https://github.com/ourtigarage">ourtigarage</a> organisation', &:contributed_out_of_org),
  Badge.new('novelist', 'The novelist', 'Wrote more than 100 words in a Pull Request\'s description', &:contribution_with_100_words),
  Badge.new('taciturn', 'The taciturn', 'Submitted a Pull Request with no description', &:contribution_with_no_word),
  Badge.new('pirate', 'The mighty pirate', 'A lawless pirate who submitted Pull Requests to his own repositories. Cheater...', &:contribution_to_own_repos),
  Badge.new('crap', 'The smelly code', 'Has a Pull Request marked as invalid. Probably some bad smelling code', &:invalid_contribs)
].freeze

# The leaderboard root class, where the magic happens
class Leaderboard
  # Initialize the leaderboard for the given event date and participant file URL
  def initialize(event_date, participants_file_url)
    @file_uri = participants_file_url
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
    # query_filter = "is:pr #{authors} created:#{@event_date} -label:invalid"
    query_filter = "#{authors} created:#{@event_date}"
    @github.search_issues(query_filter)
           .items
           .each_with_object({}) do |e, acc|
             (acc[e.user.login] ||= [e.user, []])[1] << e
           end
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
      .each_with_object(data) { |e, acc| acc[e.login] ||= [e, []] }
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
