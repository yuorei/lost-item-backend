package model

import "time"

type LostItem struct {
	ID       uint      `json:"id"`
	Kinds    []string  `json:"tags" binding:"required"`
	Comment  string    `json:"note"`
	ImageURL string    `json:"pic" binding:"required"`
	Location Location  `json:"location" binding:"required"`
	FindTime time.Time `json:"date" time_format:"2006-01-02T15:04:05Z" binding:"required"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type AreaSearchQuery struct {
	Location1 Location `json:"location1"`
	Location2 Location `json:"location2"`
}

type SearchResult struct {
	Count uint       `json:"count"`
	Items []LostItem `json:"items"`
}
