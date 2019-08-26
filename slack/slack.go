package slack

import (
	"fmt"

	"github.com/hb9tf/slack"
)

type Update struct {
	Call   string
	Status string
	Emoji  string

	Source string
}

func PostUpdate(api *slack.Client, chanID string, update Update) error {
	headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("%s via %s: %s %s", update.Call, update.Source, update.Emoji, update.Status), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)
	msg := slack.MsgOptionBlocks(headerSection)
	_, _, err := api.PostMessage(chanID, msg)
	return err
}
