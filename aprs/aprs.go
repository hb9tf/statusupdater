package aprs

import (
	"fmt"
	"log"
	"math"
	"net/textproto"
	"strings"
	"time"

	aprslib "github.com/hb9tf/go-aprs"
	"github.com/hb9tf/go-aprs/aprsis"
	"github.com/hb9tf/statusupdater/geo"
	"github.com/hb9tf/statusupdater/slack"
	"github.com/hb9tf/statusupdater/user"
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

func getFilter(filter string) string {
	fltr := filter
	if fltr == "" {
		usrs := []string{"p"}
		user.Lock.RLock()
		defer user.Lock.RUnlock()
		for call := range user.List {
			usrs = append(usrs, strings.ToLower(call))
		}
		fltr = strings.Join(usrs, "/")
	}
	return fltr
}

type Source struct {
	Endpoint string
	Callsign string
	Fltr     string

	RestartInterval time.Duration

	conn       *textproto.Conn
	packetChan chan aprslib.Packet
}

func (s *Source) Name() string { return "APRS" }

func (s *Source) process(pkt aprslib.Packet, upChan chan<- slack.Update) error {
	if pkt.Position == nil {
		return nil
	}

	// Try to use icon based on symbol received in APRS message.
	icon, err := pkt.Symbol.Emoji()
	// Use icon based on SSID as a fallback.
	if icon == "" {
		var ok bool
		icon, ok = icons[pkt.Src.SSID]
		if !ok {
			icon = defaultIcon
		}
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

	var status []string
	if pkt.Comment != "" {
		status = append(status, []string{pkt.Comment, "in"}...)
	}
	loc, err := geo.Lookup(pkt.Position.Latitude, pkt.Position.Longitude)
	if err != nil {
		log.Printf("error looking up address: %s", err)
		status = append(status, []string{
			strings.Join(pos, ""),
			fmt.Sprintf("(https://aprs.fi/%s-%d)", pkt.Src.Call, pkt.Src.SSID),
		}...)
	} else {
		status = append(status, []string{
			loc.String(),
			fmt.Sprintf("(https://aprs.fi/%s-%d)", pkt.Src.Call, pkt.Src.SSID),
		}...)
	}

	upChan <- slack.Update{
		Call:   pkt.Src.Call,
		Status: strings.Join(status, " "),
		Emoji:  icon,
		Source: s.Name(),
	}
	return nil
}

func (s *Source) startAPRS() error {
	fltr := getFilter(s.Fltr)
	log.Printf("starting APRS-IS connection using APRS filter: %s\n", fltr)
	conn, err := aprsis.Connect("tcp", s.Endpoint, s.Callsign, fltr)
	if err != nil {
		return err
	}
	if s.conn != nil {
		s.conn.Close()
	}
	go func() {
		if err := aprsis.ReadPackets(conn, s.packetChan); err != nil {
			fmt.Printf("error reading packets from APRS-IS: %s\n", err)
		}
	}()
	s.conn = conn
	return nil
}

func (s *Source) Run(upChan chan<- slack.Update) error {
	pc := make(chan aprslib.Packet, 50)
	s.packetChan = pc
	go func() {
		for {
			packet := <-s.packetChan
			if err := s.process(packet, upChan); err != nil {
				log.Printf("error processing: %s\n", err)
			}
		}
	}()

	for {
		if err := s.startAPRS(); err != nil {
			log.Printf("error starting APRS connection: %s\n", err)
		}
		time.Sleep(s.RestartInterval)
	}
}
