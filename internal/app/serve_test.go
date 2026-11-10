package app

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestServeFinishesActiveRequest(t *testing.T) {
	address := freeAddress(t)
	entered := make(chan struct{})
	release := make(chan struct{})

	e := echo.New()
	e.GET("/hold", func(c *echo.Context) error {
		close(entered)
		<-release
		return c.NoContent(http.StatusNoContent)
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(ctx, e, address)
	}()
	waitForReady(t, address)

	result := make(chan int, 1)
	go func() {
		resp, err := http.Get("http://" + address + "/hold")
		if err != nil {
			result <- 0
			return
		}
		resp.Body.Close()
		result <- resp.StatusCode
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("request did not reach the handler")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
	close(release)

	select {
	case status := <-result:
		if status != http.StatusNoContent {
			t.Fatalf("status %d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("active request was dropped")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not exit")
	}
}

func freeAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func waitForReady(t *testing.T, address string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start")
}
