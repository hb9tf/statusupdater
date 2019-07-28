package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hb9tf/statusupdater/aprs"
	"github.com/hb9tf/statusupdater/slack"
	"github.com/hb9tf/statusupdater/user"
	sl "github.com/nlopes/slack"
)

var (
	// APRS specific flags.
	aprsServer = flag.String("aprs_server", "euro.aprs2.net", "ARPS IS server to connect to")
	aprsPort   = flag.Int("aprs_port", 14580, "port to connect to on APRS IS server")
	// See http://www.aprs-is.net/Connecting.aspx
	aprsCallsign = flag.String("aprs_callsign", "", "callsign to use to log in on APRS IS")
	aprsSSID     = flag.Int("aprs_ssid", 0, "SSID to use to log in to APRS IS")
	// See http://www.aprs-is.net/javAPRSFilter.aspx
	aprsFilter = flag.String("aprs_filter", "", "APRS filter")

	// Slack specific flags.
	slackToken      = flag.String("slack_token", "", "token to use to talk to slack")
	slackExpiration = flag.Duration("slack_expiration", 30*time.Minute, "duration after which the slack status expires")
	dry             = flag.Bool("dry", false, "do not post to slack channel if true")
)

type Source interface {
	Run(chan<- slack.Update) error
}

func main() {
	flag.Parse()

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

	// start processing packets and updating the slack channel
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
			if *dry {
				log.Printf("DRY: updating %q to new status: %s\n", u.Slack.RealName, update.Status)
				continue
			}
			if err := api.SetUserCustomStatusWithUser(u.Slack.ID, update.Status, update.Emoji, time.Now().Add(*slackExpiration).Unix()); err != nil {
				log.Printf("error setting status: %v\n", err)
			} else {
				log.Printf("updated %q to new status: %s\n", u.Slack.RealName, update.Status)
			}
		}
	}()

	// create all sources and run them.
	var sources []Source
	s, err := aprs.New(*aprsServer, *aprsPort, *aprsCallsign, *aprsFilter)
	if err != nil {
		log.Fatalf("unable to create APRS source: %v", err)
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
