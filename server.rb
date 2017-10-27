require 'erubis'
require 'json'
require 'sinatra'
require_relative 'lib/leaderboard'

# URL to the participant list file. Can be local or remote
PARTICIPANT_FILE = 'https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants.md'.freeze
# The date for the event in the format YYYY[-MM[-DD]]
EVENT_DATE = '2017-10'.freeze

# Initialize the leaderboard
leaderboard = Leaderboard.new EVENT_DATE, PARTICIPANT_FILE

# Set listening port from env variable, or fallback to 80 as default
set :port, ENV['PORT'] || 80
# Set the static web content directory
set :public_folder, File.dirname(__FILE__) + '/static'

get '/' do
  erb :index, locals: { leaderboard: leaderboard }
end

get '/api/members' do
  content_type :json
  leaderboard.members.to_json
end
