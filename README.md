[![Build Status](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard.svg?branch=master)](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard)

# hacktoberfest-leaderboard
[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/ourtigarage/hacktoberfest-leaderboard)

This is the leaderboard application for the Hacktoberfest summit.
The application is kept simple so you can improve it with your own pull requests to help you
contribute for Hacktoberfest.

The application is hosted on Heroku. Visit it [here](https://hacktoberfest-leaderboard.herokuapp.com/)

Happy coding!

## How to test locally
The application is written in `Ruby`, using the [Sinatra](http://www.sinatrarb.com/) framework.
> Need to learn Ruby ? Visit [Rubymonk](https://rubymonk.com/)
### Setup dev environment
First, you need to install a `ruby` interpreter, alongside with the `gem` ruby package management tool.
Visit [Ruby language website](https://www.ruby-lang.org) for more details.

You'll probably need an editor too. [Notepad++](https://notepad-plus-plus.org/) is a simple alternative, [Visual Studio Code](https://code.visualstudio.com/) is a more advanced one.

Download `bundler` by running
```bash
    gem install bundler
```
> If you're running behind a proxy, you'll need to set both environment variables `HTTP_PROXY` and `HTTPS_PROXY`

### Running the app locally
On first run, you need to execute
```bash
    bundle install
```

Then start the application by running
```bash
    bundle exec rake run
```
then browse to http://localhost

### Running tests
In order to run unit tests, run
```bash
    bundle exec rake
```

### Useful documents
* [Sinatra usage](http://www.sinatrarb.com/intro.html) (Web microframework)
* [Octokit usage](http://www.rubydoc.info/gems/octokit/) (github api library)