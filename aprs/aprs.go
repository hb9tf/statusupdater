package aprs

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/hb9tf/statusupdater/slack"
	"github.com/hb9tf/statusupdater/user"
	aprslib "github.com/pd0mz/go-aprs"
	"github.com/pd0mz/go-aprs/aprsis"
)

var (
	/* Source: http://www.aprs.org/aprs11/SSIDs.txt
	-0 Your primary station usually fixed and message capable
	-1 generic additional station, digi, mobile, wx, etc
	-2 generic additional station, digi, mobile, wx, etc
	-3 generic additional station, digi, mobile, wx, etc
	-4 generic additional station, digi, mobile, wx, etc
	-5 Other networks (Dstar, Iphones, Androids, Blackberry's etc)
	-6 Special activity, Satellite ops, camping or 6 meters, etc
	-7 walkie talkies, HT's or other human portable
	-8 boats, sailboats, RV's or second main mobile
	-9 Primary Mobile (usually message capable)
	-10 internet, Igates, echolink, winlink, AVRS, APRN, etc
	-11 balloons, aircraft, spacecraft, etc
	-12 APRStt, DTMF, RFID, devices, one-way trackers*, etc
	-13 Weather stations
	-14 Truckers or generally full time drivers
	-15 generic additional station, digi, mobile, wx, etc
	*/
	defaultIcon = ":pager:"
	icons       = map[int]string{
		0:  ":house:",
		2:  ":car:",
		6:  ":rocket:",
		7:  ":runner:",
		8:  ":boat:",
		9:  ":pager:",
		11: ":airplane:",
		13: ":cloud:",
		14: ":truck:",
	}
)

func New(server string, port int, callsign, filter string) (*Source, error) {
	fltr := filter
	if fltr == "" {
		usrs := []string{"p"}
		user.Lock.RLock()
		defer user.Lock.RUnlock()
		for call := range user.List {
			usrs = append(usrs, strings.ToLower(call))
		}
		fltr = strings.Join(usrs, "/")
		log.Printf("updated APRS filter to: %s", fltr)
	}

	// create a connection to APRS feed
	addr := fmt.Sprintf("%s:%d", server, port)
	client, err := aprsis.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if err = client.Login(aprslib.MustParseAddress(callsign), fltr); err != nil {
		return nil, err
	}
	return &Source{client}, nil
}

type Source struct {
	client *aprsis.APRSIS
}

func (s *Source) Name() string { return "APRS" }

func (s *Source) process(pkt aprslib.Packet, upChan chan<- slack.Update) error {
	if pkt.Position == nil {
		//log.Printf("unable to process position without position: %+v", pkt)
		return nil
	}
	icon, ok := icons[pkt.Src.SSID]
	if !ok {
		icon = defaultIcon
	}
	pos := []string{
		fmt.Sprintf("%.5f", math.Abs(pkt.Position.Latitude)),
		"N",
		" ",
		fmt.Sprintf("%.5f", math.Abs(pkt.Position.Longitude)),
		"E",
	}
	if pkt.Position.Latitude < 0 {
		pos[1] = "S"
	}
	if pkt.Position.Longitude < 0 {
		pos[4] = "W"
	}
	upChan <- slack.Update{
		Call:   pkt.Src.Call,
		Status: fmt.Sprintf("%s %s (https://aprs.fi/%s-%d)", pkt.Comment, strings.Join(pos, ""), pkt.Src.Call, pkt.Src.SSID),
		Emoji:  icon,
		Source: s.Name(),
	}
	return nil
}

func (s *Source) Run(upChan chan<- slack.Update) error {
	packetChan := make(chan aprslib.Packet, 50)
	go func() {
		for {
			packet := <-packetChan
			if err := s.process(packet, upChan); err != nil {
				log.Printf("error processing: %v\n", err)
			}
		}
	}()

	// read from APRS feed
	if err := s.client.ReadPackets(packetChan); err != nil {
		return err
	}
	return nil
}
