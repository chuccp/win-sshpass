package main

import (
	"flag"
	"fmt"
	"os"

	sshpass "github.com/chuccp/win-sshpass"
)

// keygenGlobalFlags carries the values of global flags that the keygen
// subcommand understands. These are populated from the main flag set and may
// be overridden by flags that appear after the "keygen" subcommand.
type keygenGlobalFlags struct {
	algo    string
	comment string
	outPath string
}

// runKeygen implements the `keygen` subcommand. It generates an SSH key pair
// locally and saves it to disk. Deployment of the public key to a remote
// server is intentionally NOT automated — server environments vary widely and
// automatic deployment can cause issues. Users should deploy the public key
// manually.
//
// Both "flags-before-subcommand" (win-sshpass -out k keygen) and
// "flags-after-subcommand" (win-sshpass keygen -out k) styles are supported.
// Flags after the subcommand take precedence.
func runKeygen(args []string, gf keygenGlobalFlags) {
	fs := flag.NewFlagSet("keygen", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: win-sshpass keygen [options]")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fmt.Fprintln(os.Stderr, "  -algo <ed25519|rsa>    key algorithm (default: ed25519)")
		fmt.Fprintln(os.Stderr, "  -comment <string>      comment for the public key (default: user@hostname)")
		fmt.Fprintln(os.Stderr, "  -out <path>            output path for private key (default: ~/.ssh/id_ed25519 or ~/.ssh/id_rsa)")
		fmt.Fprintln(os.Stderr, "\nThe public key is written to <path>.pub alongside the private key.")
		fmt.Fprintln(os.Stderr, "Deploy the public key to the server, e.g.:")
		fmt.Fprintln(os.Stderr, `  cat ~/.ssh/id_ed25519.pub | ssh user@host "cat >> ~/.ssh/authorized_keys"`)
	}
	algo := fs.String("algo", gf.algo, "key algorithm (ed25519 or rsa)")
	comment := fs.String("comment", gf.comment, "public key comment")
	outPath := fs.String("out", gf.outPath, "output path for private key")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	jsonSetCommand(fmt.Sprintf("keygen -algo %s -out %s", *algo, *outPath))

	algoVal, err := sshpass.ParseKeyAlgorithm(*algo)
	if err != nil {
		fatalError("%v", err)
	}

	// resolve comment: explicit > user@hostname
	cmt := *comment
	if cmt == "" {
		cmt = defaultKeyComment()
	}

	// generate the key pair
	pair, err := sshpass.GenerateKeyPair(algoVal, cmt)
	if err != nil {
		fatalError("%v", err)
	}

	// determine output path
	actualPath := *outPath
	if actualPath == "" {
		actualPath = sshpass.DefaultKeyPath(algoVal)
	}

	// save to disk
	if err := sshpass.SaveKeyPair(pair, actualPath); err != nil {
		fatalError("%v", err)
	}

	// build summary text
	pubKeyLine := ""
	if pair != nil {
		pubKeyLine = string(pair.PublicKey)
	}

	// JSON mode: output structured result
	if jsonEnabled() {
		summary := fmt.Sprintf("Generated %s key pair:\n  Private key: %s\n  Public key:  %s.pub\n  Public key line: %s",
			algoVal, actualPath, actualPath, pubKeyLine)
		jsonSuccess(summary)
		return
	}

	// Normal mode: print summary to stderr
	fmt.Fprintf(os.Stderr, "Generated %s key pair:\n", algoVal)
	fmt.Fprintf(os.Stderr, "  Private key: %s\n", actualPath)
	fmt.Fprintf(os.Stderr, "  Public key:  %s.pub\n", actualPath)
	displayPub := pubKeyLine
	if len(displayPub) > 80 {
		displayPub = displayPub[:80] + "..."
	}
	fmt.Fprintf(os.Stderr, "  Public key line: %s", displayPub)
	fmt.Fprintf(os.Stderr, "\nTo enable password-less login, deploy the public key to the server:\n")
	fmt.Fprintf(os.Stderr, `  cat %s.pub | ssh user@host "cat >> ~/.ssh/authorized_keys"`+"\n", actualPath)
}

// defaultKeyComment builds a default comment string for the generated key
// using the current user and hostname.
func defaultKeyComment() string {
	user := currentUser()
	if name := currentHostname(); name != "" {
		return fmt.Sprintf("%s@%s", user, name)
	}
	return user
}

// currentUser returns the current OS username, or "user" as a fallback.
func currentUser() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	if u := os.Getenv("USERNAME"); u != "" {
		return u
	}
	return "user"
}

// currentHostname returns the machine hostname, or "" if it cannot be determined.
func currentHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return ""
	}
	return host
}
