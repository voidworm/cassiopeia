package models

type Scenario struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CampaignID *int   `json:"campaignId,omitempty"`
	PlayCount  int    `json:"playCount"`
}
