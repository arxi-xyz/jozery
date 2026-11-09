package entity

import (
	"net"
	"strings"
)

func DeviceName(userAgent string) string {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return "Unknown device"
	}

	browser := browserName(ua)
	osName := osName(ua)
	switch {
	case browser != "" && osName != "":
		return browser + " on " + osName
	case browser != "":
		return browser
	case osName != "":
		return osName
	default:
		return "Unknown device"
	}
}

func LocationLabel(ip string) (string, bool) {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return "", true
	}
	if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() || parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return "Local", true
	}
	return "", false
}

func browserName(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/"):
		return "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera"):
		return "Opera"
	case strings.Contains(ua, "Chrome/") || strings.Contains(ua, "CriOS/"):
		return "Chrome"
	case strings.Contains(ua, "Firefox/") || strings.Contains(ua, "FxiOS/"):
		return "Firefox"
	case strings.Contains(ua, "Safari/"):
		return "Safari"
	case strings.Contains(ua, "Postman"):
		return "Postman"
	case strings.Contains(ua, "curl/"):
		return "curl"
	default:
		return ""
	}
}

func osName(ua string) string {
	switch {
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		return "iOS"
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "Mac OS"):
		return "macOS"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return ""
	}
}
