package user

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hb9tf/slack"
)

var (
	List = map[string]User{}
	Lock sync.RWMutex

	callRE = regexp.MustCompile(`(\w{2}\d\w{2,3})`)
)

type User struct {
	Slack            slack.User
	LastUpdate       time.Time
	LastUpdateSource string
}

func UpdateList(api *slack.Client) {
	log.Println("fetching list of users")
	slackUsers, err := api.GetUsers()
	if err != nil {
		log.Printf("unable to get list of users: %s\n", err)
		return
	}

	Lock.Lock()
	defer Lock.Unlock()
	// add / update users
	for _, su := range slackUsers {
		if su.Deleted {
			log.Printf("ignoring deleted user: %s / %s\n", su.Name, su.RealName)
			continue
		}
		calls := callRE.FindAllString(strings.ToUpper(su.RealName), -1)
		if len(calls) == 0 {
			calls = callRE.FindAllString(strings.ToUpper(su.Profile.DisplayName), -1)
		}
		if len(calls) == 0 {

			log.Printf("ignoring user without callsign: %s / %s\n", su.Name, su.RealName)
			continue
		}
		for _, call := range calls {
			usr, ok := List[call]
			if ok {
				log.Printf("user updated: %s with callsign %s\n", su.RealName, call)
			} else {
				usr = User{}
				log.Printf("user added: %s with callsign %s\n", su.RealName, call)
			}
			usr.Slack = su
			List[call] = usr
		}
	}
	// remove users
	for call := range List {
		found := false
		for _, su := range slackUsers {
			if strings.Contains(strings.ToUpper(su.RealName), call) {
				found = true
				break
			}
			if strings.Contains(strings.ToUpper(su.Profile.DisplayName), call) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		log.Printf("user removed: callsign %s (no longer listed in slack)\n", call)
		delete(List, call)
	}
}
