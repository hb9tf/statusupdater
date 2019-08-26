package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	sl "github.com/hb9tf/slack"
	"github.com/hb9tf/statusupdater/aprs"
	"github.com/hb9tf/statusupdater/slack"
	"github.com/hb9tf/statusupdater/user"
)

var (
	// APRS specific flags.
	aprsServer   = flag.String("aprs_server", "euro.aprs2.net", "ARPS IS server to connect to")
	aprsPort     = flag.Int("aprs_port", 14580, "port to connect to on APRS IS server")
	aprsCallsign = flag.String("aprs_callsign", "NOCALL", "callsign to use to log in on APRS IS") // See http://www.aprs-is.net/Connecting.aspx
	aprsFilter   = flag.String("aprs_filter", "", "APRS filter")                                  // See http://www.aprs-is.net/javAPRSFilter.aspx

	// Slack specific flags.
	slackToken          = flag.String("slack_token", "", "token to use to talk to slack")
	slackChan           = flag.String("slack_channel", "", "ID of the slack channel to post updates in")
	slackExpiration     = flag.Duration("slack_expiration", 10*time.Minute, "duration after which the slack status expires")
	slackUpdateInterval = flag.Duration("slack_update_interval", time.Minute, "do not update slack status more often than this per user")
	dry                 = flag.Bool("dry", false, "do not post to slack channel if true")
)

type Source interface {
	Name() string
	Run(chan<- slack.Update) error
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// create new slack client
	api := sl.New(*slackToken)

	// populate the user list for the first time and then spin off a regular update
	user.UpdateList(api)
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			user.UpdateList(api)
		}
	}()

	// start processing packets and updating slack
	upChan := make(chan slack.Update, 5)
	go func() {
		for {
			update := <-upChan

			user.Lock.RLock()
			u, ok := user.List[strings.ToUpper(update.Call)]
			user.Lock.RUnlock()
			if !ok {
				log.Printf("no user found for %s", update.Call)
				continue
			}

			if !u.LastUpdate.IsZero() && u.LastUpdate.Add(*slackUpdateInterval).After(time.Now()) {
				log.Printf("not updating user %q already again (last update %s by %s)", update.Call, u.LastUpdate, u.LastUpdateSource)
				continue
			}

			u.LastUpdate = time.Now()
			u.LastUpdateSource = update.Source
			user.Lock.Lock()
			user.List[strings.ToUpper(update.Call)] = u
			user.Lock.Unlock()

			if *dry {
				log.Printf("DRY: updating %q to new status: %s\n", u.Slack.RealName, update.Status)
				continue
			}
			// Update slack status for user.
			if err := api.SetUserCustomStatusWithUser(u.Slack.ID, update.Status, update.Emoji, time.Now().Add(*slackExpiration).Unix()); err != nil {
				log.Printf("error setting status: %s\n", err)
			} else {
				log.Printf("updated %q to new status: %s\n", u.Slack.RealName, update.Status)
			}
			// Post update in channel if one was specified.
			if *slackChan == "" {
				continue
			}
			if err := slack.PostUpdate(api, *slackChan, update); err != nil {
				log.Printf("failed to post message to channel: %s", err)
			}
		}
	}()

	// create all sources and run them.
	var sources []Source
	s, err := aprs.New(*aprsServer, *aprsPort, strings.ToUpper(*aprsCallsign), *aprsFilter)
	if err != nil {
		log.Fatalf("unable to create APRS source: %s", err)
	}
	sources = append(sources, s)

	var wg sync.WaitGroup
	for _, s := range sources {
		wg.Add(1)
		go func() {
			s.Run(upChan)
		}()
	}
	wg.Wait()
}
