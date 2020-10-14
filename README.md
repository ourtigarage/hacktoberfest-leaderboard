[![Build Status](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard.svg?branch=master)](https://travis-ci.org/ourtigarage/hacktoberfest-leaderboard)

# hacktoberfest-leaderboard
[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/ourtigarage/hacktoberfest-leaderboard)

This is the leaderboard application for the Hacktoberfest summit.
The application is kept simple so you can improve it with your own pull requests to help you
contribute for Hacktoberfest.

The application is hosted on Heroku. Visit it [here](https://hacktoberfest-leaderboard.herokuapp.com/)

Happy coding!

## Configuring the app
Configuration is exclusively done by setting environment variables:
* `PORT` : The port to bind HTTP to. Default to `8080`
* `GH_TOKEN` : The token to authenticated to github. By default, no token is used, so API calls are not authenticated.
* `EVENT_DATE` : The date to restrict contribution search to. It must follows the github search date format (more details [here](https://help.github.com/articles/understanding-the-search-syntax/#query-for-dates)). Default value is `>=2005` which basically fetch everything without any restriction
* `PARTICIPANTS_FILE` : The URI or file path to the file containing the participants' github usernames. See [this file](https://raw.githubusercontent.com/ourtigarage/hacktoberfest-leaderboard/master/tests/resources/participants.md) for an example of how to format that file
* **[Deprecated]** `OBJECTIVE` : Number of pull requests to make in order to complete the challenge. Default value is `4`

> If the app is running behind a proxy, you'll need to set both environment variables `HTTP_PROXY` and `HTTPS_PROXY` before running it

### Building & Running the app
Build the application by running
```bash
    go build .
```

Then run
```bash
    ./leaderboard
    # or
    ./leaderboard.exe
```
then browse to http://localhost:8080
