require 'erubis'
require 'json'
require 'sinatra'
require_relative 'lib/leaderboard'

require 'byebug' if ENV['APP_ENV'] == 'development'

# URL to the participant list file. Can be local or remote
PARTICIPANTS_FILE = (ENV['PARTICIPANTS_FILE'] || 'https://raw.githubusercontent.com/ourtigarage/hacktoberfest-summit/master/participants/2017.md').freeze
# The date for the event in the format github date
# search format. See https://help.github.com/articles/understanding-the-search-syntax/#query-for-dates
# Default to ">=2005", since gti was released in 2005
EVENT_DATE = (ENV['EVENT_DATE'] || '>=2005').freeze

Member.objective = (ENV['OBJECTIVE'].to_i || 5).freeze

# Initialize the leaderboard
leaderboard = Leaderboard.new EVENT_DATE, PARTICIPANTS_FILE

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
