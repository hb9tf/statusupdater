# Status Updater

## Caveats

There are some caveats with this:

*  It does not do geo reverse lookups yet so the status will just contain the raw coordinates for now...

*  It only supports APRS at the moment but the idea is to add Wires-X and others too.

*  There's only basic throttling per user - if someone uses their callsign for more than one device at a time, there might be some confusing status :)

*  It requires the Slack Real Name or Display Name to contain the callsign(s).

*  The APRS filter (if left empty) is the list of callsigns found in Slack - the filter is only populated once though and changes will not be reflected. If this is a problem, consider specifying -aprs_filter directly.

*  Obviously only run one instance in the same workspace at the time...

## Installing

        $ go get -u github.com/hb9tf/statusupdater

### AuthN/AuthZ

Go to https://api.slack.com/apps and create a new app.

Give it the following permissions:

*  Send messages as StatusUpdater: `chat:write:bot`
*  Access your workspace's profile information: `users:read`
*  Access user's profile and workspace profile fields: `users.profile:read`
*  Modify user's profile: `users.profile:write`

Copy the OAuth access token.

Install the app to your workspace (as prompted by slack).

## Running

Run in dry-mode first to make sure it behaves nicely:

```
$ go run src/github.com/hb9tf/statusupdater/updater.go -aprs_callsign=<Your Callsign> -slack_token=<OAuth token> -dry >&2
```

The real deal:

```
$ go run src/github.com/hb9tf/statusupdater/updater.go -aprs_callsign=<Your Callsign> -slack_token=<OAuth token> >&2
```

### Docker

Build Docker image:

`docker build -t updater .`

Run Docker image:

`docker run --rm -e "TOKEN=<OAuth token>" -e "CALLSIGN=<Your Callsign>" updater`
