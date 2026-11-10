package usecases

import (
	"context"
	"errors"
	"time"

	"jozery/internal/domains/authentication/db"
	"jozery/internal/domains/authentication/entity"
)

type SessionInfo struct {
	ID        int64     `json:"id"`
	Device    string    `json:"device"`
	Location  string    `json:"location"`
	IP        string    `json:"ip"`
	LastLogin time.Time `json:"last_login"`
	Current   bool      `json:"current"`
}

func (s *Service) Sessions(ctx context.Context, accessToken string) ([]SessionInfo, error) {
	current, err := s.currentSession(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListSessions(ctx, current.UserID)
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionInfo, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, sessionInfo(row, row.ID == current.ID))
	}
	return sessions, nil
}

func (s *Service) RevokeSession(ctx context.Context, accessToken string, sessionID int64) error {
	current, err := s.currentSession(ctx, accessToken)
	if err != nil {
		return err
	}
	if sessionID == current.ID {
		return ErrCurrentSession
	}
	err = s.repo.RevokeSession(ctx, current.UserID, sessionID)
	if errors.Is(err, db.ErrNotFound) {
		return ErrSessionNotFound
	}
	return err
}

func (s *Service) RevokeOtherSessions(ctx context.Context, accessToken string) error {
	current, err := s.currentSession(ctx, accessToken)
	if err != nil {
		return err
	}
	return s.repo.RevokeOtherSessions(ctx, current.UserID, current.ID)
}

func sessionInfo(token entity.Token, current bool) SessionInfo {
	device := token.Device
	if device == "" {
		device = entity.DeviceName(token.UserAgent)
	}
	location := token.Location
	if location == "" {
		if label, local := entity.LocationLabel(token.IP); local {
			location = label
		}
	}
	return SessionInfo{
		ID:        token.ID,
		Device:    device,
		Location:  location,
		IP:        token.IP,
		LastLogin: token.LoggedInAt,
		Current:   current,
	}
}
