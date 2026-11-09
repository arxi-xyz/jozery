package usecases

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"jozery/internal/authentication/entity"
)

type Client struct {
	IP        string
	UserAgent string
}

type sessionMeta struct {
	IP         string
	UserAgent  string
	Device     string
	Location   string
	LoggedInAt time.Time
}

var (
	geoLookupURL = "https://ipwho.is/"
	geoClient    = &http.Client{Timeout: time.Second}
)

func describeClient(ctx context.Context, client Client) sessionMeta {
	return sessionMeta{
		IP:         strings.TrimSpace(client.IP),
		UserAgent:  client.UserAgent,
		Device:     entity.DeviceName(client.UserAgent),
		Location:   locate(ctx, client.IP),
		LoggedInAt: time.Now(),
	}
}

func locate(ctx context.Context, ip string) string {
	label, done := entity.LocationLabel(ip)
	if done {
		return label
	}
	host := net.ParseIP(strings.TrimSpace(ip))
	if host == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, geoLookupURL+url.PathEscape(host.String()), nil)
	if err != nil {
		return ""
	}
	resp, err := geoClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}

	var geo struct {
		Success bool   `json:"success"`
		City    string `json:"city"`
		Country string `json:"country"`
	}
	if err := json.Unmarshal(body, &geo); err != nil || !geo.Success {
		return ""
	}
	city := strings.TrimSpace(geo.City)
	country := strings.TrimSpace(geo.Country)
	switch {
	case city != "" && country != "":
		return city + ", " + country
	case country != "":
		return country
	default:
		return city
	}
}
