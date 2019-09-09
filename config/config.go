package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Reserves []Reserve `json:"reserves"`
	SendTo []*string `json:"mailTo"`
	MailFrom string `json:"mailFrom"`
	MailDebug string `json:"mailDebug"`
	AWSSESConfig AWSSES `json:"awsSes"`
}

type Reserve struct {
	Title string `json:"title"`
	IsShowRestaurant bool `json:"isShowRestaurant"`
	IsWeeklyCheck bool `json:"isWeeklyCheck"`
	API APIConfig `json:"api"`
	Link LinkConfig `json:"link"`
	Endpoint string `json:"endPoint"`
	Pagelink string `json:"pageLink"`
	WeeklyEndpoint string `json:"weeklyEndpoint"`
}

type APIConfig struct {
	Xhr string `json:"_xhr"`
	CommodityCD string `json:"commodityCD"`
	ContentsCd string `json:"contentsCd"`
	NameCd string `json:"nameCd"`
	UseDate string `json:"useDate"`
	AdultNum string `json:"adultNum"`
	ChildNum string `json:"childNum"`
	ChildAgeInform string `json:"childAgeInform"`
	WheelchairCount string `json:"wheelchairCount"`
	StretcherCount string `json:"stretcherCount"`
	ShowId string `json:"showId"`
	SeatRank string `json:"seatRank"`
}

type LinkConfig struct {
	ContentsCd string `json:"contentsCd"`
	NameCd string `json:"nameCd"`
	UseDate string `json:"useDate"`
	AdultNum string `json:"adultNum"`
	ChildNum string `json:"childNum"`
	ChildAgeInform string `json:"childAgeInform"`
	WheelchairCount string `json:"wheelchairCount"`
	StretcherCount string `json:"stretcherCount"`
	ShowId string `json:"showId"`

	MealDivInform string `json:"mealDivInform"`
	RestaurantTypeInform string `json:"restaurantTypeInform"`
	RestaurantNameCd string `json:"restaurantNameCd"`
	ShowWay string `json:"showWay"`
	ReservationStatus string `json:"reservationStatus"`
}

type AWSSES struct {
	AWSSESKey string `json:"awsSESKey"`
	AWSSESSecret string `json:"awsSESSecret"`
}



func Parse(filename string) (*Config, error) {
	var conf Config
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(f, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
