require 'sinatra'
require 'net/http'
require 'octokit'

PARTICIPANT_FILE = 'https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants.md'.freeze
EVENT_DATE = '2017-10'.freeze

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
    members_file.lines
                .map(&:strip)
                .map { |l| l.match(/^\* .*@([a-zA-Z0-9]+).*$/) }
                .reject(&:nil?)
                .map { |m| m[1] }
                .uniq
  end

  # Build a list of members with additional data from GitHub
  def members
    members_names.map { |m| get_user_from_github m }
                 .reject(&:nil?)
                 .map { |u| Member.new u, self }
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

  def members_file
    # Get member file at https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants.md
    uri = @file_uri
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = uri.scheme == 'https'
    resp = http.get(uri.path)
    resp.value
    resp.body
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
    @leaderboard = leaderboard
    @contribs = nil
  end

  # List member's contributions for the event month
  def month_contributions
    # Query contributions only if not already done
    @contribs ||= @leaderboard.member_contributions(@username)
  end

  # Check if the user has completed the challenge
  def challenge_complete?
    month_contributions.size >= 4
  end

  # Returns the completion percentage
  def challenge_completion
    [100, ((contributions_count.to_f / 4.0)*100).to_i].min
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

# Initialize the leaderboard
leaderboard = Leaderboard.new EVENT_DATE, PARTICIPANT_FILE

# Set listening port from env variable, or fallback to 80 as default
set :port, (ENV['PORT'] || 80).to_i
# Set the static web content directory
set :public_folder, File.dirname(__FILE__) + '/static'

get '/' do
  erb :index, locals: { leaderboard: leaderboard }
end

get '/api/members' do
  content_type :json
  leaderboard.members.to_json
end
