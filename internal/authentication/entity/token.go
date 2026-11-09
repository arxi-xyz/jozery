package entity

import "time"

type Token struct {
	UserID           int64
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
