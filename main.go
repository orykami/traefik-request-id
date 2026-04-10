package traefik_request_id

import (
	"context"
	"crypto/rand"
	"net/http"
)

const defaultHeaderName = "X-Request-ID"

// Config holds the plugin configuration.
type Config struct {
	HeaderName        string `json:"headerName,omitempty"`
	SetResponseHeader bool   `json:"setResponseHeader,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName:        defaultHeaderName,
		SetResponseHeader: true,
	}
}

type requestID struct {
	next              http.Handler
	headerName        string
	setResponseHeader bool
}

// New creates a new plugin instance.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	headerName := config.HeaderName
	if headerName == "" {
		headerName = defaultHeaderName
	}

	return &requestID{
		next:              next,
		headerName:        headerName,
		setResponseHeader: config.SetResponseHeader,
	}, nil
}

func (r *requestID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := req.Header.Get(r.headerName)
	if !isValidUUIDv4(id) {
		id = newUUID()
	}

	req.Header.Set(r.headerName, id)

	r.next.ServeHTTP(rw, req)

	if r.setResponseHeader {
		rw.Header().Set(r.headerName, id)
	}
}

// --- UUID utilities ---

const hexTable = "0123456789abcdef"

// newUUID generates a UUID v4 (RFC 4122) without external dependencies.
func newUUID() string {
	var b [16]byte

	_, _ = rand.Read(b[:])

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

// isValidUUIDv4 checks UUID v4 format without allocations or regex.
func isValidUUIDv4(s string) bool {
	if len(s) != 36 {
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	if s[14] != '4' {
		return false
	}
	switch s[19] {
	case '8', '9', 'a', 'b', 'A', 'B':
	default:
		return false
	}
	return isHex(s[0:8]) && isHex(s[9:13]) && isHex(s[14:18]) && isHex(s[19:23]) && isHex(s[24:36])
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
