# frozen_string_literal: true

# Class reprensenting a badge a player can earn
class Badge
  attr_reader :short, :title, :description

  def initialize(short, title, description, &block)
    @short = short
    @title = title
    @description = description
    @block = block
  end

  def earned_by?(player)
    # Return false if no evaluation predicate has been provided
    times_earned_by(player).positive?
  end

  def times_earned_by(player)
    # Return 0 if no evaluation predicate has been provided
    result = @block&.call(player) || 0
    result == true ? 1 : result
  end
end
