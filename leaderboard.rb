require 'sinatra'
require 'net/http'
require 'github_api'

PARTICIPANT_FILE = 'https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants.md'.freeze
EVENT_DATE = '2017-10-01'.freeze

# The leaderboard root class
class Leaderboard
  def initialize
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
    query = "created:>#{EVENT_DATE} author:#{username}"
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
    uri = URI(PARTICIPANT_FILE)
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

# Initialize the laderboard
leaderboard = Leaderboard.new

# Set listenning port from env variable, or fallback to 80 as default
set :port, (ENV['PORT'] || 80).to_i
# Set the static web content directory
set :public_folder, File.dirname(__FILE__) + '/static'

get '/' do
  erb :index, locals: { leaderboard: leaderboard }
end

get '/api/members' do
  leaderboard.members.to_json
end
