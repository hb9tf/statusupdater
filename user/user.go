package user

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nlopes/slack"
)

var (
	List = map[string]User{}
	Lock sync.RWMutex

	callRE = regexp.MustCompile(`((HB3|HB9)[\w]+)`)
)

type User struct {
	Slack            slack.User
	LastUpdate       *time.Time
	LastUpdateSource string
}

func UpdateList(api *slack.Client) {
	log.Println("fetching list of users")
	slackUsers, err := api.GetUsers()
	if err != nil {
		log.Printf("unable to get list of users: %v\n", err)
		return
	}

	Lock.Lock()
	defer Lock.Unlock()
	// add / update users
	for _, su := range slackUsers {
		call := callRE.FindString(strings.ToUpper(su.RealName))
		if call == "" {
			log.Printf("ignoring user without callsign: %s / %s\n", su.Name, su.RealName)
			continue
		}
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
	// remove users
	for call := range List {
		found := false
		for _, su := range slackUsers {
			if strings.Contains(strings.ToUpper(su.RealName), call) {
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
