# frozen_string_literal: true

# A contest member from the participant list in landing page
class Member
  attr_reader :username, :avatar, :profile, :contributions, :invalids, :issues

  def self.objective=(target)
    @@objective = target
  end

  def self.objective
    @@objective
  end

  # Construct a user using the data fetched from GitHub
  def initialize(github_user, contributions)
    @username = github_user.login
    @avatar = github_user.avatar_url
    @profile = github_user.html_url
    @invalids = []
    @contributions = []
    @issues = []
    add_contributions(contributions)
  end

  # Check if the user has completed the challenge
  def challenge_complete?
    contributions.size >= @@objective
  end

  # Returns the completion percentage
  def challenge_completion
    [100, ((contributions_count.to_f / @@objective.to_f) * 100).to_i].min
  end

  # Count the number of valid contributions
  def contributions_count
    contributions.size
  end

  def contributed_to_snake
    contributions.count { |c| c.repository_url == SNAKE_URL }
  end

  def contributed_to_leaderboard
    contributions.count { |c| c.repository_url == LEADERBOARD_URL }
  end

  def ten_contributions?
    contributions.size >= 10
  end

  def contributed_out_of_org
    contributions.count do |c|
      !c.repository_url.start_with?(ORG_REPOS_URL) &&
        !c.repository_url.start_with?("#{BASE_REPOS_URL}/#{@username}")
    end
  end

  def contribution_with_100_words
    contributions.count { |c| c.body.split(' ').count >= 100 }
  end

  def contribution_with_no_word
    contributions.count { |c| c.body.strip.empty? }
  end

  def contribution_to_own_repos
    contributions.count do |c|
      c.repository_url.start_with? "#{BASE_REPOS_URL}/#{@username}"
    end
  end

  def invalid_contribs
    invalids.size
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

  private

  def add_contributions(contributions)
    contributions.each do |contrib|
      if !contrib.pull_request
        @issues << contrib
      elsif contrib.labels.any? { |l| l.name == 'invalid' }
        @invalids << contrib
      else
        @contributions << contrib
      end
    end
  end
end
