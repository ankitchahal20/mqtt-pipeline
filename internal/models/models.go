package models

type SpeedData struct {
	Speed *int `json:"speed"`
}

type Email struct {
	Email string `json:"email,omitempty"`
}
