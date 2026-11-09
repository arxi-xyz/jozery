package entity

import "time"

type Token struct {
	ID               int64
	UserID           int64
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	IP               string
	UserAgent        string
	Device           string
	Location         string
	LoggedInAt       time.Time
}
