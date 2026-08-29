package models

type Investigator struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ClassID   *int   `json:"classId,omitempty"`
	PlayCount int    `json:"playCount"`
}
