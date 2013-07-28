package http

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
)

var localhost = &net.TCPAddr{
	IP:   net.IP{127, 0, 0, 1},
	Port: 0, // os assigned
}

type server struct {
	*testing.T
	net.Listener
}

// Shutdown should be called to terminate this server.
func (s *server) Shutdown() {
	s.Listener.Close()
}

// Root returns a http URL for the root of this server.
func (s *server) Root() string {
	return fmt.Sprintf("http://%s", s.Listener.Addr().String())
}

// starts a new net/http http server
func newServer(t *testing.T, mux *http.ServeMux) *server {
	l, err := net.ListenTCP("tcp4", localhost)
	if err != nil {
		t.Fatal(err)
	}
	// /404 is not handled, generating a 404
	go func() {
		if err := http.Serve(l, mux); err != nil {
			// t.Error(err)
		}
	}()
	return &server{t, l}
}

func sameErr(a, b error) bool {
	if a != nil && b != nil {
		return a.Error() == b.Error()
	}
	return a == b
}

func TestInternalHttpServer(t *testing.T) {
	newServer(t, nil).Shutdown()
}

var a string

func init() {
	for i := 0; i < 1024; i++ {
		a += "aaaaaaaa"
	}
}

var getTests = []struct {
	path     string
	expected string
	err      error
}{
	{"/200", "OK", nil},
	{"/404", "", errors.New("404 Not Found")},
	// {"/a", a, nil},	// triggers chunked encoding
}

func TestGet(t *testing.T) {
	s := newServer(t, stdmux())
	defer s.Shutdown()
	for _, tt := range getTests {
		url := s.Root() + tt.path
		var b bytes.Buffer
		n, err := Get(&b, url)
		if actual := b.String(); actual != tt.expected || n != int64(len(tt.expected)) || !sameErr(err, tt.err) {
			t.Errorf("Get(%q): expected %q %v, got %q %v", tt.path, tt.expected, tt.err, actual, err)
		}
	}
}
