// Package sshpass provides a reusable SSH client SDK with password and
// private-key authentication, command execution, interactive shells (with
// rz/sz file-transfer support), and SFTP-based file transfer.
//
// A typical usage:
//
//	cfg := sshpass.NewConfig()
//	cfg.Host = "example.com"
//	cfg.User = "root"
//	cfg.Password = "secret"
//
//	client, err := sshpass.NewClient(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	if err := client.Exec("uname -a"); err != nil {
//	    log.Fatal(err)
//	}
//
// The behavior of a Client can be customized through options, for example to
// redirect I/O streams, report transfer progress via a callback, or supply a
// file dialog for rz/sz:
//
//	client, err := sshpass.NewClient(cfg,
//	    sshpass.WithStdin(in),
//	    sshpass.WithStdout(out),
//	    sshpass.WithProgress(func(desc string, sent, total int64) { ... }),
//	)
package sshpass
