package sshpass

import (
	"fmt"
	"io"
	"log/slog"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// agentAuthMethod connects to the local ssh-agent and returns an ssh.AuthMethod
// using the keys available in the agent. The returned io.Closer must be kept
// alive for the duration of the SSH handshake and closed afterwards.
func agentAuthMethod() (ssh.AuthMethod, io.Closer, error) {
	conn, err := agentDial()
	if err != nil {
		return nil, nil, err
	}

	agentClient := agent.NewClient(conn)
	signers, err := agentClient.Signers()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to get signers from agent: %w", err)
	}
	if len(signers) == 0 {
		conn.Close()
		return nil, nil, fmt.Errorf("ssh-agent has no keys loaded")
	}

	return ssh.PublicKeys(signers...), conn, nil
}

// setupAgentForwarding connects to the local ssh-agent and registers a
// forwarding handler on the SSH client. This allows the remote server to
// use the local agent for further SSH connections. The returned io.Closer
// must be kept alive for the duration of the SSH session.
func setupAgentForwarding(sshClient *ssh.Client, logger *slog.Logger) (io.Closer, error) {
	conn, err := agentDial()
	if err != nil {
		return nil, err
	}

	agentClient := agent.NewClient(conn)
	if err := agent.ForwardToAgent(sshClient, agentClient); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to forward agent: %w", err)
	}

	logger.Info("ssh-agent forwarding enabled")
	return conn, nil
}

// requestAgentForwarding sends an agent-forwarding request for the given
// session. This must be called after setupAgentForwarding and before
// starting the shell or command.
func requestAgentForwarding(session *ssh.Session) {
	// Errors here are non-fatal — forwarding is a convenience feature.
	_ = agent.RequestAgentForwarding(session)
}
