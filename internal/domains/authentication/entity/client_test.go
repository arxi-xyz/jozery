package entity

import "testing"

func TestDeviceName(t *testing.T) {
	cases := []struct {
		ua   string
		want string
	}{
		{"", "Unknown device"},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", "Chrome on Linux"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", "Edge on Windows"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1", "Safari on iOS"},
		{"Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", "Chrome on Android"},
		{"PostmanRuntime/7.36.0", "Postman"},
	}
	for _, tc := range cases {
		if got := DeviceName(tc.ua); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.ua, got, tc.want)
		}
	}
}

func TestLocationLabel(t *testing.T) {
	label, local := LocationLabel("127.0.0.1")
	if label != "Local" || !local {
		t.Fatalf("loopback: %q %v", label, local)
	}
	label, local = LocationLabel("8.8.8.8")
	if label != "" || local {
		t.Fatalf("public: %q %v", label, local)
	}
}
