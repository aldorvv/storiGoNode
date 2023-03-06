package models

import "time"

type Account struct {
	ID   int    `json:"id"`
	File string `json:"file"`
}

type Move struct {
	ID        int       `json:"id"`
	Kind      int       `json:"kind"`
	Quantity  float64   `json:"quantity"`
	Date      time.Time `json:"date"`
	AccountID int       `json:"account"`
}
