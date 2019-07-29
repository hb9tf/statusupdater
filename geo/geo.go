package geo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
)

const (
	lookupServerTmpl = "https://nominatim.openstreetmap.org/reverse.php?format=json&lat=%f&lon=%f"
)

type Location struct {
	Class       string  `json:"class,omitempty"`
	DisplayName string  `json:"display_name,omitempty"`
	Importance  float64 `json:"importance,omitempty"`
	Latitude    float64 `json:"lat,string,omitempty"`
	Longitude   float64 `json:"lon,string,omitempty"`
	OSMID       int64   `json:"osm_id,omitempty"`
	OSMType     string  `json:"osm_type,omitempty"`
	PlaceID     int64   `json:"place_id,omitempty"`
	Type        string  `json:"type,omitempty"`
	Address     *struct {
		Village       string `json:"village,omitempty"`
		Town          string `json:"town,omitempty"`
		City          string `json:"city,omitempty"`
		CityDistrict  string `json:"city_district,omitempty"`
		Continent     string `json:"continent,omitempty"`
		Country       string `json:"country,omitempty"`
		CountryCode   string `json:"country_code,omitempty"`
		County        string `json:"county,omitempty"`
		Hamlet        string `json:"hamlet,omitempty"`
		HouseNumber   string `json:"house_number,omitempty"`
		Pedestrian    string `json:"pedestrian,omitempty"`
		Neighbourhood string `json:"neighbourhood,omitempty"`
		PostCode      string `json:"postcode,omitempty"`
		Road          string `json:"road,omitempty"`
		State         string `json:"state,omitempty"`
		StateDistrict string `json:"state_district,omitempty"`
		Suburb        string `json:"suburb,omitempty"`
	} `json:"address"`
	BoundingBox []string `json:"boundingbox,omitempty"`
}

func (l *Location) String() string {
	if l.Address == nil {
		pos := []string{
			fmt.Sprintf("%.5f", math.Abs(l.Latitude)),
			"N",
			" ",
			fmt.Sprintf("%.5f", math.Abs(l.Longitude)),
			"E",
		}
		if l.Latitude < 0 {
			pos[1] = "S"
		}
		if l.Longitude < 0 {
			pos[4] = "W"
		}

		return strings.Join(pos, "")
	}

	var loc []string
	if l.Address.City != "" {
		loc = append(loc, l.Address.City)
	} else if l.Address.Town != "" {
		loc = append(loc, l.Address.Town)
	} else if l.Address.Village != "" {
		loc = append(loc, l.Address.Village)
	} else {
		return l.DisplayName
	}

	if l.Address.CountryCode != "" {
		loc = append(loc, strings.ToUpper(l.Address.CountryCode))
	} else if l.Address.Country != "" {
		loc = append(loc, l.Address.Country)
	}

	return strings.Join(loc, " ")
}

func Lookup(lat, lon float64) (*Location, error) {
	url := fmt.Sprintf(lookupServerTmpl, lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loc Location
	if err := json.Unmarshal(body, &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}
