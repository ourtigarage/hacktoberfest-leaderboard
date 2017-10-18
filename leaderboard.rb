require 'sinatra'
require 'net/http'
require 'github_api'

PARTICIPANT_FILE = 'https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants.md'.freeze
EVENT_DATE = '2017-10'.freeze

# The leaderboard root class, where the magic happen
class Leaderboard
  # Initialize the leaderboard for the given event date and participant file URL
  def initialize(event_date, participants_file_url)
    @file_uri = URI(participants_file_url)
    @event_date = event_date
    # Conect to github using a token from env variable.
    # If no token is set, no problem it will still work,
    # but with limited rate for API calls
    @github = Github.new oauth_token: ENV['GH_TOKEN']
  end

  # Retrieve the list of participants from github page
  def members_names
    # Extract usernames from file
    members_file.lines
                .map(&:strip)
                .map { |l| l.match(/^\* .*@([a-zA-Z0-9]+).*$/) }
                .reject(&:nil?)
                .map { |m| m[1] }
                .uniq
  end

  # Build a list of members with additional data from github
  def members
    members_names.map { |m| get_user_from_github m }
                 .reject(&:nil?)
                 .map { |u| Member.new u }
  end

  # Retrieve list of user's pull requests from github
  def member_contributions(username)
    query = "created:#{@event_date} author:#{username} -label:invalid"
    contribs = @github.search.issues(query)
    contribs.body.items.reject { |i| i.pull_request.nil? }
  end

  private

  def get_user_from_github(username)
    @github.users.get user: username
  rescue Github::Error::NotFound
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

  # Construct a user from data fetched from github
  def initialize(github_user)
    @username = github_user.login
    @fullname = github_user.name
    @avatar = github_user.avatar_url
    @profile = github_user.html_url
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

# Set listenning port from env variable, or fallback to 80 as default
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
