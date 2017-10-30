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
    return false if !@block
    result = @block.call(player)
    result == false ? false : result == true || result > 0
  end

  def times_earned_by(player)
    # Return 0 if no evaluation predicate has been provided
    return 0 if !@block
    result = @block.call(player)
    return 0 if result == false
    result == true ? 1 : result
  end
end
