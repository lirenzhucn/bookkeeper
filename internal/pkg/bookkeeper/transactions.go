package bookkeeper

import "time"

type Transaction struct {
	Id           int       `json:"Id"`
	Date         time.Time `json:"Date"`
	Desc         string    `json:"Desc"`
	OriginalDesc string    `json:"OriginalDesc"`
	Amount       float32   `json:"Amount"`
	Type         string    `json:"Type"`
	Category     string    `json:"Category"`
	AccountId    int       `json:"AccountId"`
	Notes        string    `json:"Notes"`
}
