# Status Updater

## Caveats

There are some caveats with this:

*  It uses the github.com/nlopes/slack library but that doesn't support modifying other users' slack status out of the box hence you need to patch it (patch provided). I also created https://github.com/nlopes/slack/issues/562 to track this.

*  It does not do geo reverse lookups yet so the status will just contain the raw coordinates for now...

*  It only supports APRS at the moment but the idea is to add Wires-X and others too.

*  There's no rate-limiting or other smart-ness - if someone uses their callsign for more than one device at a time, there might be some confusing status :)

*  It requires the Slack Real Names to contain the callsigns. Format the Real Names like this: First Last (Callsign).

*  Obviously only run one instance in the same workspace at the time...

## Installing

        $ go get -u github.com/hb9tf/statusupdater

### Patching

Patch `github.com/nlopes/slack/users.go` with `slack.users.go.diff` - see https://github.com/nlopes/slack/issues/562 for more details.

Note that the Dockerfile does that automagically when building.

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
