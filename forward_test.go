package sshpass

import (
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

type forwardingTestClient struct {
	dial func(network, address string) (net.Conn, error)
}

func (c forwardingTestClient) Dial(network, address string) (net.Conn, error) {
	return c.dial(network, address)
}

func (forwardingTestClient) Listen(string, string) (net.Listener, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestForwarderPreservesResponseAfterClientHalfClose(t *testing.T) {
	upstreamLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer upstreamLn.Close()

	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer proxyLn.Close()

	f := &Forwarder{
		client: forwardingTestClient{dial: func(network, address string) (net.Conn, error) {
			return net.Dial(network, address)
		}},
		target:    upstreamLn.Addr().String(),
		direction: "local",
	}

	clientConn, err := net.Dial("tcp", proxyLn.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()
	proxyConn, err := proxyLn.Accept()
	if err != nil {
		t.Fatal(err)
	}

	handleDone := make(chan struct{})
	go func() {
		f.handle(proxyConn)
		close(handleDone)
	}()

	serverDone := make(chan error, 1)
	go func() {
		upstream, err := upstreamLn.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer upstream.Close()
		request, err := io.ReadAll(upstream)
		if err != nil {
			serverDone <- err
			return
		}
		if string(request) != "request" {
			serverDone <- fmt.Errorf("request = %q, want request", request)
			return
		}
		if _, err := upstream.Write([]byte("response")); err != nil {
			serverDone <- err
			return
		}
		serverDone <- upstream.(*net.TCPConn).CloseWrite()
	}()

	if _, err := clientConn.Write([]byte("request")); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.(*net.TCPConn).CloseWrite(); err != nil {
		t.Fatal(err)
	}
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	response, err := io.ReadAll(clientConn)
	if err != nil {
		t.Fatalf("failed to read forwarded response: %v", err)
	}
	if string(response) != "response" {
		t.Errorf("response = %q, want response", response)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("upstream failed: %v", err)
	}

	select {
	case <-handleDone:
	case <-time.After(3 * time.Second):
		t.Fatal("forwarder handler did not finish")
	}
}

func TestForwarderCloseTerminatesActiveConnections(t *testing.T) {
	targetLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer targetLn.Close()

	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	f := &Forwarder{
		listener:  proxyLn,
		target:    targetLn.Addr().String(),
		direction: "remote",
		done:      make(chan struct{}),
	}
	go f.acceptLoop()

	targetConnCh := make(chan net.Conn, 1)
	go func() {
		conn, err := targetLn.Accept()
		if err == nil {
			targetConnCh <- conn
		}
	}()

	clientConn, err := net.Dial("tcp", proxyLn.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	var targetConn net.Conn
	select {
	case targetConn = <-targetConnCh:
		defer targetConn.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("target did not accept the forwarded connection")
	}
	if _, err := clientConn.Write([]byte("keep-open")); err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if err := f.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	if _, err := clientConn.Read(buf); err == nil {
		t.Fatal("active client connection remained open after Forwarder.Close")
	} else {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatal("active client connection was not closed by Forwarder.Close")
		}
	}
}
