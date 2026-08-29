package models

import "time"

type Session struct {
	ID         int       `json:"id"`
	ScenarioID *int      `json:"scenarioId,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

type SessionPlayer struct {
	SessionID      int `json:"sessionId"`
	InvestigatorID int `json:"investigatorId"`
}
