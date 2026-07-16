package sshpass

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// proxyDial establishes a TCP connection to address (host:port) through the
// proxy specified by proxyURL. Supported schemes:
//   - socks5://[user:pass@]host:port  — SOCKS5, DNS resolved locally
//   - socks5h://[user:pass@]host:port — SOCKS5, DNS resolved by proxy
//   - socks4://[user@]host:port       — SOCKS4
//   - http://[user:pass@]host:port    — HTTP CONNECT
//   - https://[user:pass@]host:port   — HTTPS CONNECT (TLS to proxy)
//
// timeout is the dial timeout in seconds (0 = no limit). Authentication
// credentials are taken from the URL userinfo.
func proxyDial(proxyURL, address string, timeout int) (net.Conn, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "socks5", "socks5h", "socks4":
		return socksDial(u, address, timeout)
	case "http", "https":
		return httpConnectDial(u, scheme, address, timeout)
	default:
		return nil, fmt.Errorf("unsupported proxy scheme %q (use socks5, socks5h, socks4, http, or https)", scheme)
	}
}

// socksDial handles SOCKS4/SOCKS5 proxies. SOCKS5 uses golang.org/x/net/proxy;
// SOCKS4 is implemented inline (the protocol is trivial and x/net/proxy has
// no SOCKS4 support).
func socksDial(u *url.URL, address string, timeout int) (net.Conn, error) {
	scheme := strings.ToLower(u.Scheme)
	proxyAddr := u.Host
	if !strings.Contains(proxyAddr, ":") {
		proxyAddr = net.JoinHostPort(proxyAddr, "1080")
	}

	if scheme == "socks4" {
		return socks4Dial(proxyAddr, u, address, timeout)
	}
	return socks5Dial(proxyAddr, u, address, timeout)
}

// socks5Dial creates a SOCKS5 dialer via golang.org/x/net/proxy. The timeout
// covers both the TCP connection to the proxy AND the SOCKS5 protocol handshake
// (auth + CONNECT request/response), so a proxy that accepts the TCP connection
// but never responds to SOCKS5 negotiation won't hang forever.
func socks5Dial(proxyAddr string, u *url.URL, address string, timeout int) (net.Conn, error) {
	var forward proxy.Dialer = &net.Dialer{}
	if timeout > 0 {
		forward = &net.Dialer{Timeout: time.Duration(timeout) * time.Second}
	}

	var auth *proxy.Auth
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		auth = &proxy.Auth{User: user, Password: pass}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, forward)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	// proxy.Dialer.Dial does not accept a context, so when a timeout is
	// configured we run the dial in a goroutine and race it against a timer.
	// This ensures a SOCKS5 proxy that hangs during negotiation (after the TCP
	// connection is established) is still subject to the deadline.
	type dialResult struct {
		conn net.Conn
		err  error
	}
	if timeout <= 0 {
		conn, err := dialer.Dial("tcp", address)
		if err != nil {
			return nil, fmt.Errorf("SOCKS5 proxy connection failed: %w", err)
		}
		return conn, nil
	}

	resultCh := make(chan dialResult, 1)
	go func() {
		conn, err := dialer.Dial("tcp", address)
		resultCh <- dialResult{conn, err}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil {
			return nil, fmt.Errorf("SOCKS5 proxy connection failed: %w", res.err)
		}
		return res.conn, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		// Best-effort: the goroutine may still complete and write to resultCh
		// (buffered, so no leak); the connection it created (if any) will be
		// GC'd or closed by the proxy dialer's internal cleanup.
		return nil, fmt.Errorf("SOCKS5 proxy connection timed out after %ds", timeout)
	}
}

// socks4Dial implements the SOCKS4/SOCKS4A CONNECT command. When the target
// host is not an IP literal, SOCKS4A is used (host name sent to the proxy).
func socks4Dial(proxyAddr string, u *url.URL, address string, timeout int) (net.Conn, error) {
	var d net.Dialer
	if timeout > 0 {
		d.Timeout = time.Duration(timeout) * time.Second
	}
	conn, err := d.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SOCKS4 proxy %s: %w", proxyAddr, err)
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("invalid target address %q: %w", address, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		conn.Close()
		return nil, fmt.Errorf("invalid target port %q", portStr)
	}

	// SOCKS4 userid (optional, from proxy URL username).
	userid := ""
	if u.User != nil {
		userid = u.User.Username()
	}

	ip := net.ParseIP(host)
	if ip != nil {
		// SOCKS4: destination is a 4-byte IPv4 address.
		ip4 := ip.To4()
		if ip4 == nil {
			conn.Close()
			return nil, fmt.Errorf("SOCKS4 does not support IPv6 target %s", host)
		}
		req := make([]byte, 0, 8+len(userid)+1)
		req = append(req, 0x04, 0x01) // VN=4, CD=1 (CONNECT)
		req = append(req, byte(port>>8), byte(port))
		req = append(req, ip4...)
		req = append(req, userid...)
		req = append(req, 0x00)
		if _, err := conn.Write(req); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to send SOCKS4 request: %w", err)
		}
	} else {
		// SOCKS4A: destination is a hostname. Use IP 0.0.0.x (x != 0).
		req := make([]byte, 0, 8+len(userid)+1+len(host)+1)
		req = append(req, 0x04, 0x01)
		req = append(req, byte(port>>8), byte(port))
		req = append(req, 0x00, 0x00, 0x00, 0x01) // invalid IP → 4A hostname mode
		req = append(req, userid...)
		req = append(req, 0x00)
		req = append(req, host...)
		req = append(req, 0x00)
		if _, err := conn.Write(req); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to send SOCKS4A request: %w", err)
		}
	}

	// Read 8-byte reply: VN(0) CD(status) DSTPORT(2) DSTIP(4).
	var resp [8]byte
	if _, err := io.ReadFull(conn, resp[:]); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading SOCKS4 reply: %w", err)
	}
	if resp[1] != 0x5a {
		conn.Close()
		return nil, fmt.Errorf("SOCKS4 proxy rejected connection (status 0x%02x)", resp[1])
	}
	return conn, nil
}

// httpConnectDial handles HTTP/HTTPS CONNECT proxies. For https it wraps the
// underlying TCP connection to the proxy in TLS before issuing CONNECT.
func httpConnectDial(u *url.URL, scheme, address string, timeout int) (net.Conn, error) {
	proxyHost := u.Hostname()
	proxyPort := u.Port()
	if proxyPort == "" {
		if scheme == "https" {
			proxyPort = "443"
		} else {
			proxyPort = "80"
		}
	}
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)

	var d net.Dialer
	if timeout > 0 {
		d.Timeout = time.Duration(timeout) * time.Second
	}

	conn, err := d.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP proxy %s: %w", proxyAddr, err)
	}

	// For https proxies, wrap the connection in TLS before sending CONNECT.
	if scheme == "https" {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: proxyHost})
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("TLS handshake with proxy failed: %w", err)
		}
		conn = tlsConn
	}

	// Build the CONNECT request with optional Basic auth.
	var authHeader string
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		creds := user + ":" + pass
		authHeader = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(creds)) + "\r\n"
	}
	connectReq := "CONNECT " + address + " HTTP/1.1\r\nHost: " + address + "\r\n" + authHeader + "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT to proxy: %w", err)
	}

	// Read and validate the proxy's HTTP response status line + headers.
	// Use a bufio.Reader for efficient line reads; any bytes it buffers beyond
	// the headers (e.g. the start of the SSH handshake pushed by the server)
	// must be preserved — see bufferedConn below.
	br := bufio.NewReader(conn)
	statusLine, err := br.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading proxy response: %w", err)
	}
	// Expect "HTTP/1.1 200 Connection established\r\n" (or similar).
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		conn.Close()
		return nil, fmt.Errorf("malformed proxy response: %q", strings.TrimSpace(statusLine))
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("malformed proxy status code: %q", parts[1])
	}
	if code != 200 {
		conn.Close()
		return nil, fmt.Errorf("HTTP proxy returned %d %s", code, strings.TrimSpace(strings.Join(parts[2:], " ")))
	}

	// Consume remaining headers until blank line.
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("error reading proxy headers: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}

	// If the bufio.Reader has buffered bytes beyond the headers (the server
	// may have already started sending SSH handshake data), wrap the connection
	// so those bytes are served first before reading fresh data from conn.
	if br.Buffered() > 0 {
		return &bufferedConn{r: br, Conn: conn}, nil
	}
	return conn, nil
}

// bufferedConn wraps a net.Conn with a bufio.Reader so that any bytes already
// buffered in the reader are returned first on Read, before delegating to the
// underlying connection. This is necessary after reading HTTP CONNECT response
// headers with a bufio.Reader: the proxy (or the SSH server behind it) may push
// data immediately after the headers, and that data would be stuck in the
// bufio buffer if we returned the raw conn.
type bufferedConn struct {
	r *bufio.Reader
	net.Conn
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}
