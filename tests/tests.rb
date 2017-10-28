require 'test/unit'
require_relative '../lib/badge'
require_relative '../lib/member'
require_relative '../lib/leaderboard'

# Test case for leaderboard
class TestLeaderboard < Test::Unit::TestCase
  # This tests nothing for now
  def test_members_discovery
    md_file = "#{File.dirname(__FILE__)}/resources/participants.md"
    leaderboard = Leaderboard.new '2017-10', md_file
    names = leaderboard.members_names
    assert_equal 6, names.size
    assert_equal %w[MisterX22 Blimy phsym nerfox22 Kirack2040 WonderWolski],
                 names
  end
end
