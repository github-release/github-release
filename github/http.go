package github

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Github's HTTP endpoint for asset uploads is flaky. Often it will drop a
// TCP connection *. The HTTP POST will just hang indefinitely. To solve
// this without using absolute deadlines (which would be hard to predict as
// it depends on the size of the asset and the speed of the users'
// connection), we use a modified http.Client which resets a timeout every
// time the kernel accepts a read/write on the socket. Doing so basically
// creates a sort of "inactivity watcher".
//
// We can't use the http.Client.Timeout field, as that's an absolute timeout
// which doesn't get reset whenever there's some activity.
//
// * At least that's what I think is happening, I have never observed this
//   myself since I never upload big assets, see issue
//   http://github.com/aktau/github-release/issues/26.

// HTTPReadWriteTimeout is the read/write timeout after which connections
// will be closed.
const HTTPReadWriteTimeout = 1 * time.Minute

// dialer for use by the transport below, initialized like the net/http
// dialer.
var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
	DualStack: true,
}

// transport for use by the http.Client below.
//
// TODO: enable HTTP/2 transport as documented in net/http.
var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
		return newWatchdogConn(dialer.DialContext(ctx, netw, addr))
	},
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ResponseHeaderTimeout: 1 * time.Minute,
}

// client is an HTTP client suited for use with GitHub. It includes watchdog
// functionality that will break of an upload/download attempt after
// HTTPReadWriteTimeout.
var client = &http.Client{Transport: transport}

// watchdogConn wraps a net.Conn, on every read and write it sets a timeout.
// If no bytes are read or written in that time, the connection is closed.
type watchdogConn struct {
	net.Conn
	timeout time.Duration // The amount of time to wait between reads/writes on the connection before cancelling it.
}

func newWatchdogConn(conn net.Conn, err error) (net.Conn, error) {
	return &watchdogConn{Conn: conn, timeout: HTTPReadWriteTimeout}, err
}

func (c *watchdogConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(b)
}

func (c *watchdogConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.Conn.Write(b)
}
