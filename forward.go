package sshpass

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// Forwarder manages an active port-forwarding tunnel. It is returned by
// Client.LocalForward and Client.RemoteForward. The accept loop runs in a
// background goroutine; call Close to stop the tunnel and release resources.
type Forwarder struct {
	listener  net.Listener
	client    sshClientConn
	target    string // address to dial for each accepted connection
	direction string // "local" or "remote"
	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

// sshClientConn is the subset of *ssh.Client used by Forwarder. It is
// extracted as an interface so forward.go does not need to import
// golang.org/x/crypto/ssh directly.
type sshClientConn interface {
	Dial(network, address string) (net.Conn, error)
	Listen(network, address string) (net.Listener, error)
}

// LocalForward creates a local port-forwarding tunnel. It listens on
// localAddr (e.g. "127.0.0.1:8080") and forwards each accepted connection to
// remoteAddr (e.g. "db.internal:3306") through the SSH connection.
//
// The returned Forwarder runs its accept loop in a background goroutine.
// Call Close to stop the tunnel.
func (c *Client) LocalForward(localAddr, remoteAddr string) (*Forwarder, error) {
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}
	f := &Forwarder{
		listener:  ln,
		client:    c.sshClient,
		target:    remoteAddr,
		direction: "local",
		done:      make(chan struct{}),
	}
	go f.acceptLoop()
	return f, nil
}

// RemoteForward creates a remote port-forwarding tunnel. It asks the SSH
// server to listen on remoteAddr (e.g. "0.0.0.0:9090") and forwards each
// accepted connection to localAddr (e.g. "127.0.0.1:8080") on the local
// machine.
func (c *Client) RemoteForward(remoteAddr, localAddr string) (*Forwarder, error) {
	ln, err := c.sshClient.Listen("tcp", remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on remote %s: %w", remoteAddr, err)
	}
	f := &Forwarder{
		listener:  ln,
		client:    c.sshClient,
		target:    localAddr,
		direction: "remote",
		done:      make(chan struct{}),
	}
	go f.acceptLoop()
	return f, nil
}

// acceptLoop accepts connections in a loop until the listener is closed.
func (f *Forwarder) acceptLoop() {
	defer close(f.done)
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return // listener closed
		}
		go f.handle(conn)
	}
}

// handle forwards a single connection bidirectionally.
func (f *Forwarder) handle(conn net.Conn) {
	defer conn.Close()

	var upstream net.Conn
	var err error
	if f.direction == "local" {
		// Local forward: dial the remote target through the SSH tunnel.
		upstream, err = f.client.Dial("tcp", f.target)
	} else {
		// Remote forward: dial the local target directly.
		upstream, err = net.Dial("tcp", f.target)
	}
	if err != nil {
		return
	}
	defer upstream.Close()

	// Bidirectional copy. When either direction finishes, close both
	// connections by returning (deferred closes handle cleanup).
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(upstream, conn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(conn, upstream)
		done <- struct{}{}
	}()
	<-done
}

// Close stops the forwarding tunnel and releases the listener. It is safe to
// call multiple times.
func (f *Forwarder) Close() error {
	f.closeOnce.Do(func() {
		f.closeErr = f.listener.Close()
	})
	return f.closeErr
}

// Wait blocks until the forwarding tunnel is closed (by Close or a listener
// error).
func (f *Forwarder) Wait() error {
	<-f.done
	return nil
}
