package server

import (
	"bytes"
	"log/slog"
	"net"
	"testing"
)

func init() {
	srv, err := NewServer(":1055", 1024)
	if err != nil {
		slog.Error("error creating server")
		return
	}

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("error starting server")
			return
		}
	}()
}

func TestServerResponse(t *testing.T) {
	tt := []struct {
		name    string
		payload []byte
		want    []byte
	}{
		{
			"Send a simple payload",
			[]byte("ping"),
			[]byte("ping"),
		},
		{
			"Send another simple payload",
			[]byte("pong"),
			[]byte("pong"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			udpAddr, err := net.ResolveUDPAddr("udp", ":1055")
			if err != nil {
				t.Error("error resolving UDP address: ", err)
			}

			conn, err := net.DialUDP("udp", nil, udpAddr)
			if err != nil {
				t.Error("error connecting to server: ", err)
			}
			defer conn.Close()

			if _, err := conn.Write(tc.payload); err != nil {
				t.Error("error writing to server: ", err)
			}

			out := make([]byte, len(tc.want))
			if _, err := conn.Read(out); err != nil {
				t.Error("error reading from server: ", err)
			}
			if !bytes.Equal(out, tc.want) {
				t.Errorf("got %b, want %b", out, tc.want)
			}
		})
	}
}
