require 'concurrent'
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
                   .uniq
                   .to_a
  end

  # Build a list of members with additional data from GitHub
  def members
    members_names.map { |m| Future.execute { get_user_from_github m } }
                 .each { |f| raise f.reason if !f.value && f.rejected? }
                 .map(&:value)
                 .reject(&:nil?)
                 .map { |u| Future.execute { Member.new(u, self) } }
                 .each { |f| raise f.reason if !f.value && f.rejected? }
                 .map(&:value)
  end

  # Retrieve list of user's pull requests from github
  def member_contributions(username)
    query = "created:#{@event_date} author:#{username} -label:invalid"
    contribs = @github.search_issues(query)
    contribs.items.reject { |i| i.pull_request.nil? }
  end

  private

  def get_user_from_github(username)
    @github.user username
  rescue Octokit::NotFound
    nil
  end
end

# A contest member from the participant list in landing page
class Member
  attr_reader :username, :fullname, :avatar, :profile

  # Construct a user using the data fetched from GitHub
  def initialize(github_user, leaderboard)
    @username = github_user.login
    @fullname = github_user.name
    @avatar = github_user.avatar_url
    @profile = github_user.html_url
    # @leaderboard = leaderboard
    @contribs = leaderboard.member_contributions(@username)
  end

  # List member's contributions for the event month
  def month_contributions
    # Query contributions only if not already done
    # @contribs ||= @leaderboard.member_contributions(@username)
    @contribs
  end

  # Check if the user has completed the challenge
  def challenge_complete?
    month_contributions.size >= 4
  end

  # Returns the completion percentage
  def challenge_completion
    [100, ((contributions_count.to_f / 4.0) * 100).to_i].min
  end

  # Count the number of valid contributions
  def contributions_count
    month_contributions.size
  end

  def to_json(*_opts)
    {
      username: @username,
      fullname: @fulname,
      avatar: @avatar,
      profile: @profile
    }.to_json
  end
end
