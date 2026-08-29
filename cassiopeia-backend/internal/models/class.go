package models

type Class struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Colour *string `json:"colour,omitempty"`
}
