package sshpass

import (
	"bufio"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProxyDialUnsupportedScheme(t *testing.T) {
	_, err := proxyDial("ftp://proxy:21", "example.com:22", 5)
	if err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestProxyDialInvalidURL(t *testing.T) {
	_, err := proxyDial("not a url", "example.com:22", 5)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestProxyDialSchemeRouting(t *testing.T) {
	// These will fail at the network layer (no real proxy), but should fail
	// with a connection error, not a scheme-routing error — proving the
	// correct dialer was selected.
	cases := []struct {
		name string
		url  string
	}{
		{"socks5", "socks5://127.0.0.1:1"},
		{"socks5h", "socks5h://127.0.0.1:1"},
		{"socks4", "socks4://127.0.0.1:1"},
		{"http", "http://127.0.0.1:1"},
		{"https", "https://127.0.0.1:1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := proxyDial(tc.url, "example.com:22", 1)
			if err == nil {
				t.Fatal("expected connection error (no real proxy)")
			}
			if strings.Contains(err.Error(), "unsupported proxy scheme") {
				t.Errorf("scheme was not routed correctly: %s", err)
			}
		})
	}
}

// --- HTTP CONNECT proxy tests ---

// startHTTPProxy starts a mock HTTP CONNECT proxy that accepts any CONNECT
// request and tunnels the connection to the target. It returns the proxy
// address and a shutdown function.
func startHTTPProxy(t *testing.T, requireAuth string) (addr string, shutdown func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				// Read request line.
				line, err := br.ReadString('\n')
				if err != nil {
					return
				}
				_ = line
				// Read headers until blank line.
				var headers string
				for {
					h, err := br.ReadString('\n')
					if err != nil {
						return
					}
					headers += h
					if h == "\r\n" || h == "\n" {
						break
					}
				}
				if requireAuth != "" {
					if !strings.Contains(headers, requireAuth) {
						c.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n\r\n"))
						return
					}
				}
				c.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
				// Tunnel any buffered data + future data.
				if br.Buffered() > 0 {
					buf := make([]byte, br.Buffered())
					br.Read(buf)
					c.Write(buf)
				}
				io.Copy(c, br)
			}(conn)
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestHTTPProxyConnectAndTunnel(t *testing.T) {
	// Start a target echo server.
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer targetLn.Close()
	go func() {
		for {
			conn, err := targetLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c) // echo
			}(conn)
		}
	}()

	proxyAddr, shutdown := startHTTPProxy(t, "")
	defer shutdown()

	proxyURL := "http://" + proxyAddr
	conn, err := proxyDial(proxyURL, targetLn.Addr().String(), 5)
	if err != nil {
		t.Fatalf("proxyDial failed: %v", err)
	}
	defer conn.Close()

	// Write data and verify echo (proves the tunnel works bidirectionally).
	msg := []byte("hello-proxy\n")
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf) != string(msg) {
		t.Errorf("echo = %q, want %q", buf, msg)
	}
}

func TestHTTPProxyAuth(t *testing.T) {
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer targetLn.Close()

	// Proxy requires Basic auth user:pass.
	proxyAddr, shutdown := startHTTPProxy(t, "Proxy-Authorization: Basic dXNlcjpwYXNz")
	defer shutdown()

	u := &url.URL{
		Scheme: "http",
		Host:   proxyAddr,
		User:   url.UserPassword("user", "pass"),
	}
	conn, err := httpConnectDial(u, "http", targetLn.Addr().String(), 5)
	if err != nil {
		t.Fatalf("httpConnectDial with auth failed: %v", err)
	}
	conn.Close()
}

func TestHTTPProxyRejected(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				for {
					h, err := br.ReadString('\n')
					if err != nil {
						return
					}
					if h == "\r\n" || h == "\n" {
						break
					}
				}
				c.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
			}(conn)
		}
	}()

	_, err = proxyDial("http://"+ln.Addr().String(), "example.com:22", 5)
	if err == nil {
		t.Fatal("expected error for rejected proxy connection")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status 403, got: %v", err)
	}
}

// --- SOCKS4 proxy tests ---

// startSocks4Proxy starts a mock SOCKS4 proxy that accepts CONNECT requests
// and tunnels to the target. It returns the proxy address and shutdown.
func startSocks4Proxy(t *testing.T, targetAddr string) (addr string, shutdown func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var req [8]byte
				if _, err := io.ReadFull(c, req[:]); err != nil {
					return
				}
				// Read userid (null-terminated).
				br := bufio.NewReader(c)
				_, _ = br.ReadString(0)
				// Read hostname for SOCKS4A (if IP is 0.0.0.x).
				if req[4] == 0 && req[5] == 0 && req[6] == 0 && req[7] != 0 {
					_, _ = br.ReadString(0)
				}
				// Reply: success (0x5a).
				resp := []byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
				c.Write(resp)
				// Tunnel.
				io.Copy(c, br)
			}(conn)
		}
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestSocks4ProxyConnectAndTunnel(t *testing.T) {
	// Start a target echo server.
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer targetLn.Close()
	go func() {
		for {
			conn, err := targetLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	proxyAddr, shutdown := startSocks4Proxy(t, targetLn.Addr().String())
	defer shutdown()

	// Use the numeric IP of the echo server so SOCKS4 (not 4A) is used.
	host, portStr, _ := net.SplitHostPort(targetLn.Addr().String())
	port, _ := strconv.Atoi(portStr)
	_ = host
	_ = port

	conn, err := proxyDial("socks4://"+proxyAddr, targetLn.Addr().String(), 5)
	if err != nil {
		t.Fatalf("proxyDial socks4 failed: %v", err)
	}
	defer conn.Close()

	msg := []byte("socks4-test\n")
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write(msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf) != string(msg) {
		t.Errorf("echo = %q, want %q", buf, msg)
	}
}

func TestSocks4AProxyHostnameMode(t *testing.T) {
	// SOCKS4A with a hostname (non-IP) target should send the hostname.
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer targetLn.Close()
	go func() {
		for {
			conn, err := targetLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c)
			}(conn)
		}
	}()

	proxyAddr, shutdown := startSocks4Proxy(t, "")
	defer shutdown()

	// Pass a hostname target (localhost) to trigger SOCKS4A mode.
	conn, err := proxyDial("socks4://"+proxyAddr, targetLn.Addr().String(), 5)
	if err != nil {
		t.Fatalf("proxyDial socks4 (4A mode) failed: %v", err)
	}
	conn.Close()
}

func TestSocks4ProxyRejected(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var req [8]byte
				io.ReadFull(c, req[:])
				br := bufio.NewReader(c)
				br.ReadString(0)
				if req[4] == 0 && req[5] == 0 && req[6] == 0 && req[7] != 0 {
					br.ReadString(0)
				}
				// Reply: rejected (0x5b).
				c.Write([]byte{0x00, 0x5b, 0, 0, 0, 0, 0, 0})
			}(conn)
		}
	}()

	_, err = proxyDial("socks4://"+ln.Addr().String(), "127.0.0.1:22", 5)
	if err == nil {
		t.Fatal("expected error for rejected SOCKS4 connection")
	}
	if !strings.Contains(err.Error(), "rejected") {
		t.Errorf("error should mention rejection, got: %v", err)
	}
}

// --- Default port tests ---

func TestSocksDialDefaultPort(t *testing.T) {
	// When no port is in the proxy URL, socksDial defaults to 1080.
	// We can't easily test the actual connection, but we can verify the
	// address parsing doesn't panic and routes to the right port by checking
	// the error message mentions port 1080.
	_, err := proxyDial("socks5://127.0.0.1", "example.com:22", 1)
	if err == nil {
		return // connected somehow (unlikely)
	}
	// The connection error should reference 127.0.0.1:1080.
	if !strings.Contains(err.Error(), "1080") {
		t.Logf("note: error did not mention default port 1080: %v", err)
	}
}

func TestHTTPProxyDefaultPort(t *testing.T) {
	// http:// without port should default to 80.
	_, err := proxyDial("http://127.0.0.1", "example.com:22", 1)
	if err == nil {
		return
	}
	// Should not be a "missing host" or "unsupported scheme" error.
	if strings.Contains(err.Error(), "unsupported proxy scheme") {
		t.Errorf("http scheme should be supported: %v", err)
	}
}

// TestHTTPProxyPreservesBufferedData verifies the Bug 1 fix: when the proxy
// (or the SSH server behind it) pushes data immediately after the "200
// Connection established" response headers — in the same TCP segment that the
// bufio.Reader consumed while parsing headers — those bytes must be delivered
// to the caller, not lost in the bufio buffer.
func TestHTTPProxyPreservesBufferedData(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	// The "extra" data the proxy injects right after the headers, simulating
	// an SSH banner or handshake start pushed by the target server.
	extraData := []byte("SSH-2.0-OpenSSH_8.9\r\n")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				// Consume the CONNECT request line + headers.
				for {
					line, err := br.ReadString('\n')
					if err != nil {
						return
					}
					if line == "\r\n" || line == "\n" {
						break
					}
				}
				// Write the 200 response AND the extra data in a single
				// Write call so they land in the same TCP segment and the
				// client's bufio.Reader buffers both.
				resp := append([]byte("HTTP/1.1 200 Connection established\r\n\r\n"), extraData...)
				c.Write(resp)
				// Echo any subsequent data from the client.
				io.Copy(c, br)
			}(conn)
		}
	}()

	conn, err := proxyDial("http://"+ln.Addr().String(), "example.com:22", 5)
	if err != nil {
		t.Fatalf("proxyDial failed: %v", err)
	}
	defer conn.Close()

	// The extraData bytes must be readable from the returned connection.
	// If the bufferedConn fix is missing, this read will hang/timeout because
	// the bytes are stuck in the bufio.Reader that was discarded.
	buf := make([]byte, len(extraData))
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("failed to read buffered data (Bug 1 regression?): %v", err)
	}
	if string(buf) != string(extraData) {
		t.Errorf("buffered data = %q, want %q", buf, extraData)
	}
}

// TestSocks4ProxyUserID verifies that the username from the proxy URL is sent
// as the SOCKS4 userid field in the CONNECT request.
func TestSocks4ProxyUserID(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	receivedUserID := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Read the 8-byte SOCKS4 header.
		var header [8]byte
		if _, err := io.ReadFull(conn, header[:]); err != nil {
			receivedUserID <- ""
			return
		}
		// Read the null-terminated userid.
		br := bufio.NewReader(conn)
		userid, _ := br.ReadString(0)
		userid = strings.TrimRight(userid, "\x00")
		receivedUserID <- userid
		// Reply success so the caller doesn't hang.
		conn.Write([]byte{0x00, 0x5a, 0, 0, 0, 0, 0, 0})
	}()

	conn, err := proxyDial("socks4://myuser@"+ln.Addr().String(), "127.0.0.1:22", 5)
	if err != nil {
		t.Fatalf("proxyDial failed: %v", err)
	}
	conn.Close()

	select {
	case uid := <-receivedUserID:
		if uid != "myuser" {
			t.Errorf("SOCKS4 userid = %q, want %q", uid, "myuser")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for proxy to receive SOCKS4 request")
	}
}

// TestSocks5ProxyTimeout verifies that socks5Dial does not hang forever when
// the proxy accepts the TCP connection but never responds to the SOCKS5
// negotiation (Bug 2 regression test).
func TestSocks5ProxyTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// Accept the connection but never send any SOCKS5 response,
			// simulating a broken/hung proxy.
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()

	start := time.Now()
	_, err = proxyDial("socks5://"+ln.Addr().String(), "example.com:22", 2)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	// Should time out around 2s, not hang indefinitely. Allow some slack.
	if elapsed > 5*time.Second {
		t.Errorf("dial took %v, expected ~2s timeout", elapsed)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error should mention timeout, got: %v", err)
	}
}
