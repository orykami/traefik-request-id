package traefik_request_id

import (
	"context"
	"crypto/rand"
	"net/http"
)

const defaultHeaderName = "X-Request-ID"

// Config holds the plugin configuration.
type Config struct {
	HeaderName string `json:"headerName,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName: defaultHeaderName,
	}
}

type requestID struct {
	next       http.Handler
	headerName string
}

// New creates a new plugin instance.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	headerName := config.HeaderName
	if headerName == "" {
		headerName = defaultHeaderName
	}

	return &requestID{
		next:       next,
		headerName: headerName,
	}, nil
}

func (r *requestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := newUUID()

	req.Header.Set(r.headerName, id)

	r.next.ServeHTTP(rw, req)

	rw.Header().Set(r.headerName, id)
}

const hexTable = "0123456789abcdef"

// newUUID generates a UUID v4 (RFC 4122) without external dependencies.
func newUUID() string {
	var b [16]byte

	rand.Read(b[:])

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	var buf [36]byte
	hexEncode(buf[0:8], b[0:4])
	buf[8] = '-'
	hexEncode(buf[9:13], b[4:6])
	buf[13] = '-'
	hexEncode(buf[14:18], b[6:8])
	buf[18] = '-'
	hexEncode(buf[19:23], b[8:10])
	buf[23] = '-'
	hexEncode(buf[24:36], b[10:16])

	return string(buf[:])
}

func hexEncode(dst []byte, src []byte) {
	for i, v := range src {
		dst[i*2] = hexTable[v>>4]
		dst[i*2+1] = hexTable[v&0x0f]
	}
}
