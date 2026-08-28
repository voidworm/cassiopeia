package models

type Investigator struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	PlayCount int    `json:"playCount"`
}
