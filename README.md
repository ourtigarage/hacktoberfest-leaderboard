[![Build Status](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard.svg?branch=master)](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard)

# hacktoberfest-leaderboard
[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/ourtigarage/hacktoberfest-leaderboard)

This is the leaderboard application for the Hacktoberfest summit.
The application is kept simple so you can improve it with your own pull requests to help you
contribute for Hacktoberfest.

The application is hosted on Heroku. Visit it [here](https://hacktoberfest-leaderboard.herokuapp.com/)

Happy coding!

## How to test & run locally
The application is written in `Ruby`, using the [Sinatra](http://www.sinatrarb.com/) framework.
> Need to learn Ruby ? Visit [Rubymonk](https://rubymonk.com/)
### Setup dev environment
First, you need to install a `ruby` interpreter, alongside with the `gem` ruby package management tool.
Visit [Ruby language website](https://www.ruby-lang.org) for more details.

You'll probably need an editor too. [Notepad++](https://notepad-plus-plus.org/) is a simple alternative, [Visual Studio Code](https://code.visualstudio.com/) is a more advanced one.

> If you're running behind a proxy, you'll need to set both environment variables `HTTP_PROXY` and `HTTPS_PROXY` before going further

Download and install `bundler` by running
```bash
    gem install bundler
```

Then, go to the project directory and run
```bash
    bundle install
```

### Running tests
In order to run unit tests, run
```bash
    bundle exec rake
```

> If you're running behind a proxy, you'll need to set both environment variables `HTTP_PROXY` and `HTTPS_PROXY`
### Configuring the app
Configuration is exclusively done by setting environment variables:
* `PORT` : The port to bind HTTP to. Default to `80`
* `RACK_ENV` : Should be set to `production`. Automatically set when deploying to Heroku.
* `GH_TOKEN` : The token to authenticated to github. By default, no token is used, so API alls are not authenticated.
* `EVENT_DATE` : The date to restrict contribution search to. It must follows the github search date format (more details [here](https://help.github.com/articles/understanding-the-search-syntax/#query-for-dates)). Default value is `>=2005` which basically fetch everything without any restriction
* `PARTICIPANTS_FILE` : The URI or file path to the file containing the participants' github usernames. See [this file](https://raw.githubusercontent.com/ourtigarage/hacktoberfest-leaderboard/master/tests/resources/participants.md) for an example of how to format that file

> Again, if the app running behind a proxy, you'll need to set both environment variables `HTTP_PROXY` and `HTTPS_PROXY` before running it

### Running the app
Then start the application by running
```bash
    bundle exec rake run
```
then browse to http://localhost


### Useful documents
* [Sinatra usage](http://www.sinatrarb.com/intro.html) (Web microframework)
* [Octokit usage](http://www.rubydoc.info/gems/octokit/) (github api library)