package models

import "time"

type Advertiser struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type AdsResponse struct {
	Domain           string       `json:"domain"`
	TotalAdvertisers int          `json:"total_advertisers"`
	Advertisers      []*Advertiser `json:"advertisers"`
	Cached           bool         `json:"cached"`
	Timestamp        time.Time    `json:"timestamp"`
}
