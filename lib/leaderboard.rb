require 'concurrent'
require 'faraday-http-cache'
require 'json'
require 'octokit'
require 'open-uri'

include Concurrent

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

  # Retrieve the list of participants from GitHub page
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

  def to_json(*_opts)
    {
      username: @username,
      avatar: @avatar,
      profile: @profile
    }.to_json
  end
end
