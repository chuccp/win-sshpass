package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	sshpass "github.com/chuccp/win-sshpass"
)

// runUpdate implements the `update` subcommand: it asks GitHub for the newest
// release of win-sshpass and, unless the current build is already the newest,
// downloads the archive for this platform and replaces the running binary.
//
// Its flags are only accepted after the subcommand word (win-sshpass update
// -check); -check/-force/-version mean nothing to the ssh/scp/rsync modes, so
// unlike the keygen flags they are not registered in the global flag set.
func runUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: win-sshpass update [options]")
		fmt.Fprintln(os.Stderr, "\nDownloads the latest release from GitHub and replaces the current binary.")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fmt.Fprintln(os.Stderr, "  -check            only report whether a newer release exists")
		fmt.Fprintln(os.Stderr, "  -force            reinstall even when the current version is already the latest")
		fmt.Fprintln(os.Stderr, "  -version <tag>    install a specific release tag (e.g. v1.0.0); implies -force")
		fmt.Fprintln(os.Stderr, "  -target <path>    binary to replace (default: the running win-sshpass)")
		fmt.Fprintln(os.Stderr, "\nInstalled by scoop/winget/MSI? Update through that package manager instead.")
	}
	check := fs.Bool("check", false, "only check whether a newer release exists")
	force := fs.Bool("force", false, "reinstall even when the current version is already the latest")
	version := fs.String("version", "", "release tag to install (default: the latest release)")
	target := fs.String("target", "", "binary to replace (default: the running executable)")
	if err := fs.Parse(args); err != nil {
		// An explicit -help is a successful request, not a usage error.
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		os.Exit(2)
	}
	if fs.NArg() > 0 {
		fatalError("unexpected argument %q\nUsage: win-sshpass update [-check] [-force] [-version <tag>] [-target <path>]", fs.Arg(0))
	}

	jsonSetCommand("update")

	// Progress bar on stderr, suppressed in JSON mode so the structured result
	// stays the only thing on stdout.
	var progress sshpass.ProgressFunc
	if !jsonEnabled() {
		progress = newCLIProgress(os.Stderr).progress
	}

	verb := "Updating"
	if *check {
		verb = "Checking"
	}
	if !jsonEnabled() {
		fmt.Fprintf(os.Stderr, "%s win-sshpass %s (%s/%s)...\n", verb, sshpass.Version, runtime.GOOS, runtime.GOARCH)
	}

	result, err := sshpass.SelfUpdate(sshpass.UpdateOptions{
		Version: *version,
		// Asking for an explicit tag means "install this", even when it is not
		// newer than what is running (e.g. to undo a bad upgrade).
		Force:      *force || *version != "",
		CheckOnly:  *check,
		TargetPath: *target,
		Progress:   progress,
		Logf: func(format string, args ...any) {
			if !jsonEnabled() {
				fmt.Fprintf(os.Stderr, format+"\n", args...)
			}
		},
	})
	if err != nil {
		fatalError("update failed: %v", err)
	}

	summary := updateSummary(result, *check)
	if jsonEnabled() {
		jsonSuccess(summary)
		return
	}
	fmt.Println(summary)

	// Windows cannot delete the binary it is running from, so the parked copy
	// stays behind when this very build was replaced; it is removed by the next
	// update (see replaceExecutable). Only mention it when it is really there.
	if result.Updated && runtime.GOOS == "windows" {
		if _, err := os.Stat(result.TargetPath + ".old"); err == nil {
			fmt.Fprintf(os.Stderr, "Note: the previous build is kept as %s.old and removed by the next update.\n", result.TargetPath)
		}
	}
}

// updateSummary renders the outcome of an update for humans and for the
// stdout field of the JSON result.
func updateSummary(r *sshpass.UpdateResult, checkOnly bool) string {
	switch {
	case r.Updated && r.CurrentVersion == r.LatestVersion:
		return fmt.Sprintf("Reinstalled win-sshpass %s (%s)", r.LatestVersion, r.TargetPath)
	case r.Updated:
		return fmt.Sprintf("Updated win-sshpass %s -> %s (%s)", r.CurrentVersion, r.LatestVersion, r.TargetPath)
	case checkOnly && r.UpdateAvailable:
		return fmt.Sprintf("A newer version is available: %s -> %s\nRun 'win-sshpass update' to install it.\n%s",
			r.CurrentVersion, r.LatestVersion, r.ReleaseURL)
	case checkOnly:
		return fmt.Sprintf("win-sshpass %s is up to date (latest release: %s)", r.CurrentVersion, r.LatestVersion)
	default:
		return fmt.Sprintf("win-sshpass %s is already the latest version", r.CurrentVersion)
	}
}
