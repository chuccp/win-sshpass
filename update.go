package sshpass

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultUpdateRepo is the GitHub repository ("owner/name") that win-sshpass
// releases are published to.
const DefaultUpdateRepo = "chuccp/win-sshpass"

// releaseAPIBase is the GitHub REST API root. It is a variable so tests can
// point the update machinery at a local server.
var releaseAPIBase = "https://api.github.com"

// githubWebBase is the github.com root. Release discovery normally goes through
// it rather than the API — see latestReleaseByRedirect. It is a variable so
// tests can serve the redirects locally.
var githubWebBase = "https://github.com"

const (
	// updateAPITimeout bounds a single GitHub API request (release metadata).
	updateAPITimeout = 30 * time.Second
	// maxReleaseJSON caps how much of an API response is read.
	maxReleaseJSON = 4 << 20
	// maxBinarySize caps the size of the executable extracted from an archive.
	maxBinarySize = 256 << 20
	// minBinarySize rejects a download that is far too small to be the binary,
	// e.g. an HTML error page saved by a captive portal.
	minBinarySize = 16 << 10
)

// downloadStallTimeout aborts a transfer that stops delivering bytes, which is
// what a dead connection looks like. It is a variable so tests can shorten it.
var downloadStallTimeout = 2 * time.Minute

// ReleaseAsset is one downloadable file attached to a GitHub release.
type ReleaseAsset struct {
	// Name is the file name, e.g. "win-sshpass-v1.0.0-amd64.zip".
	Name string
	// URL is the direct download URL of the asset.
	URL string
	// Size is the asset size in bytes (0 when the API does not report it).
	Size int64
}

// ReleaseInfo describes a published GitHub release.
type ReleaseInfo struct {
	// Tag is the release tag, e.g. "v1.0.0".
	Tag string
	// HTMLURL is the release page on github.com.
	HTMLURL string
	// PublishedAt is the release publication time.
	PublishedAt time.Time
	// Assets lists the downloadable files attached to the release.
	Assets []ReleaseAsset
}

// UpdateOptions configures SelfUpdate.
type UpdateOptions struct {
	// Repo is the GitHub repository ("owner/name") to check. Defaults to
	// DefaultUpdateRepo.
	Repo string
	// Version is a specific release tag to install (e.g. "v1.0.0"). When empty
	// the newest published release is used.
	Version string
	// CheckOnly reports whether an update is available without downloading or
	// replacing anything.
	CheckOnly bool
	// Force reinstalls the release even when the current version is already up
	// to date.
	Force bool
	// TargetPath is the executable to replace. Defaults to the running
	// executable.
	TargetPath string
	// Client is the HTTP client used for the API request and the download.
	// Defaults to a client with sane timeouts that honours HTTP(S)_PROXY.
	Client *http.Client
	// Progress, when set, receives download progress as byte counts.
	Progress ProgressFunc
	// Logf, when set, receives human-readable progress messages.
	Logf func(format string, args ...any)
}

// UpdateResult reports what SelfUpdate found and did.
type UpdateResult struct {
	// CurrentVersion is the version of the running binary.
	CurrentVersion string
	// LatestVersion is the release tag that was found.
	LatestVersion string
	// UpdateAvailable reports whether LatestVersion is newer than
	// CurrentVersion.
	UpdateAvailable bool
	// Updated reports whether the executable on disk was replaced.
	Updated bool
	// TargetPath is the executable that was (or would be) replaced.
	TargetPath string
	// AssetName is the release asset that was selected.
	AssetName string
	// ReleaseURL is the release page on github.com.
	ReleaseURL string
}

// SelfUpdate checks the GitHub releases of the project and, when a newer
// release exists, downloads the archive for the current platform and replaces
// the executable in place (opts.TargetPath, default: the running binary).
//
// The replacement is staged in the target's directory and moved into place, so
// the new binary and the old one must be on the same filesystem. On Windows
// the running binary cannot be overwritten, but it can be renamed: the old
// build is moved to "<target>.old" and removed on the next update.
//
// SelfUpdate performs no rollback beyond that: a binary that was replaced once
// is not restored if a later step fails. Callers that need a guaranteed
// fallback should keep a copy of the executable themselves.
func SelfUpdate(opts UpdateOptions) (*UpdateResult, error) {
	if opts.Logf == nil {
		opts.Logf = func(string, ...any) {}
	}

	repo := opts.Repo
	if repo == "" {
		repo = DefaultUpdateRepo
	}

	target := opts.TargetPath
	if target == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("cannot locate the running executable: %w (set TargetPath to replace a specific file)", err)
		}
		target = exe
	}
	target, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target path %s: %w", target, err)
	}

	release, err := resolveRelease(repo, opts.Version, opts.Client)
	if err != nil {
		return nil, err
	}

	result := &UpdateResult{
		CurrentVersion:  Version,
		LatestVersion:   release.Tag,
		UpdateAvailable: CompareVersions(Version, release.Tag) < 0,
		TargetPath:      target,
		ReleaseURL:      release.HTMLURL,
	}
	if opts.CheckOnly || (!result.UpdateAvailable && !opts.Force) {
		return result, nil
	}

	asset, err := SelectReleaseAsset(release.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}
	result.AssetName = asset.Name

	ext := archiveExt(asset.Name)
	if ext == "" {
		return nil, fmt.Errorf("release asset %q is not a supported archive (.zip or .tar.gz)", asset.Name)
	}

	// Stage the new binary next to the target before downloading anything.
	// Replacing an executable is a rename, and a rename cannot cross
	// filesystems (EXDEV on Unix, a hard failure on Windows), so the staged
	// file must share the target's volume — and discovering that the install
	// directory is not writable should not cost the user a full download.
	staged, err := os.CreateTemp(filepath.Dir(target), ".win-sshpass-new-*")
	if err != nil {
		return nil, fmt.Errorf("cannot write to %s: %w%s", filepath.Dir(target), err, privilegedInstallHint(target))
	}
	stagedPath := staged.Name()
	staged.Close()
	defer os.Remove(stagedPath)

	tmpDir, err := os.MkdirTemp("", "win-sshpass-update-")
	if err != nil {
		return nil, fmt.Errorf("cannot create a temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "release"+ext)
	opts.Logf("downloading %s (%s)", asset.Name, humanBytes(asset.Size))
	if _, err := DownloadFile(asset.URL, archivePath, "Downloading "+asset.Name, opts.Client, opts.Progress); err != nil {
		return nil, err
	}

	if err := ExtractBinary(archivePath, stagedPath); err != nil {
		return nil, err
	}
	if err := verifyBinary(stagedPath); err != nil {
		return nil, err
	}

	opts.Logf("installing %s over %s", asset.Name, target)
	if err := replaceExecutable(target, stagedPath); err != nil {
		return nil, err
	}

	result.Updated = true
	return result, nil
}

// CompareVersions orders two version strings, with or without a leading "v",
// using semver-like rules: numeric parts compare numerically, a release sorts
// after its own pre-releases ("v1.0.0" > "v1.0.0-rc1"), and a version with no
// leading number at all (e.g. a "dev" build) sorts before every numeric
// version, so development builds see the newest release as an upgrade.
//
// It returns -1 if a is older than b, 0 if the two are equal, and +1 if a is
// newer.
func CompareVersions(a, b string) int {
	aNums, aPre, aOK := parseVersion(a)
	bNums, bPre, bOK := parseVersion(b)

	switch {
	case !aOK && !bOK:
		return 0
	case !aOK:
		return -1
	case !bOK:
		return 1
	}

	for i := 0; i < len(aNums) || i < len(bNums); i++ {
		var x, y int
		if i < len(aNums) {
			x = aNums[i]
		}
		if i < len(bNums) {
			y = bNums[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}

	// Same numbers: a final release outranks any pre-release of it.
	switch {
	case aPre == bPre:
		return 0
	case aPre == "":
		return 1
	case bPre == "":
		return -1
	case aPre < bPre:
		return -1
	default:
		return 1
	}
}

// parseVersion splits a version string into its numeric parts and its
// pre-release suffix. ok is false when the string does not start with a
// number.
func parseVersion(v string) (nums []int, pre string, ok bool) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
	if v == "" {
		return nil, "", false
	}
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		pre, v = v[i+1:], v[:i]
	}
	for _, part := range strings.Split(v, ".") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, "", false
		}
		nums = append(nums, n)
	}
	return nums, pre, len(nums) > 0
}

// FetchLatestRelease returns the newest published (non-draft, non-prerelease)
// release of repo ("owner/name") from the GitHub API. client may be nil.
//
// This always talks to api.github.com, so it is subject to its rate limit.
// SelfUpdate itself prefers the lighter github.com route; use this when you
// need the complete asset list.
func FetchLatestRelease(repo string, client *http.Client) (*ReleaseInfo, error) {
	return fetchReleaseFromAPI(repo, "", client)
}

// FetchReleaseByTag returns a specific release of repo ("owner/name") by tag
// (e.g. "v1.0.0") from the GitHub API. client may be nil.
func FetchReleaseByTag(repo, tag string, client *http.Client) (*ReleaseInfo, error) {
	return fetchReleaseFromAPI(repo, tag, client)
}

// resolveRelease returns the release to install.
//
// The github.com route is tried first: it needs no API quota (the REST API
// allows 60 unauthenticated requests per hour) and keeps working on networks
// where api.github.com is blocked or throttled while github.com is fine. The
// API is the fallback and is authoritative — it lists the real assets, so a
// release that is missing a build for this platform is reported precisely.
func resolveRelease(repo, tag string, client *http.Client) (*ReleaseInfo, error) {
	if repo == "" {
		repo = DefaultUpdateRepo
	}
	if client == nil {
		client = defaultUpdateClient(updateAPITimeout)
	}

	info, webErr := releaseFromWeb(repo, tag, client)
	if webErr == nil {
		return info, nil
	}

	apiInfo, apiErr := fetchReleaseFromAPI(repo, tag, client)
	if apiErr != nil {
		return nil, webErr
	}
	return apiInfo, nil
}

// releaseFromWeb resolves a release using nothing but github.com:
//
//	/releases/latest         answers 302 with the newest tag in Location
//	/releases/tag/<tag>      answers 200 when that tag exists
//
// The asset URL then follows from the tag, because the release workflow names
// every archive "win-sshpass-<tag>[-<goos>]-<arch>.<ext>". Keep this in step
// with .github/workflows/release.yml.
func releaseFromWeb(repo, tag string, client *http.Client) (*ReleaseInfo, error) {
	latest := tag == ""
	if latest {
		resolved, err := resolveLatestTag(repo, client)
		if err != nil {
			return nil, err
		}
		tag = resolved
	} else if err := verifyTag(repo, tag, client); err != nil {
		return nil, err
	}

	// Always download from the tag, never from /releases/latest/download/: the
	// asset name carries the tag we just compared, but that URL resolves to
	// whatever is newest when the download starts. A release published in
	// between would make it serve a different build than the one checked — or
	// 404, since the older asset is not attached to the new release. The
	// tag-pinned URL is immutable.
	downloadBase := githubWebBase + "/" + repo + "/releases/download/" + url.PathEscape(tag)

	assets := make([]ReleaseAsset, 0, 2)
	for _, suffix := range assetSuffixes(runtime.GOOS, runtime.GOARCH) {
		name := "win-sshpass-" + tag + suffix
		assets = append(assets, ReleaseAsset{Name: name, URL: downloadBase + "/" + name})
	}

	return &ReleaseInfo{
		Tag:     tag,
		HTMLURL: githubWebBase + "/" + repo + "/releases/tag/" + url.PathEscape(tag),
		Assets:  assets,
	}, nil
}

// resolveLatestTag asks github.com which release is current: /releases/latest
// answers 302 with the tag in the Location header, so one HEAD request settles
// it without spending API quota.
func resolveLatestTag(repo string, client *http.Client) (string, error) {
	resp, err := headRequest(githubWebBase+"/"+repo+"/releases/latest", client)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusFound, http.StatusMovedPermanently, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		location := resp.Header.Get("Location")
		const marker = "/releases/tag/"
		i := strings.LastIndex(location, marker)
		if i < 0 {
			return "", fmt.Errorf("GitHub answered with an unexpected redirect: %s", location)
		}
		tag, err := url.PathUnescape(location[i+len(marker):])
		if err != nil || tag == "" {
			return "", fmt.Errorf("GitHub answered with an unexpected redirect: %s", location)
		}
		return tag, nil
	case http.StatusNotFound:
		return "", fmt.Errorf("no releases published for %s yet", repo)
	default:
		return "", fmt.Errorf("github.com returned %s for the latest release", resp.Status)
	}
}

// verifyTag confirms a pinned tag exists, so a typo is reported before any
// download is attempted.
func verifyTag(repo, tag string, client *http.Client) error {
	resp, err := headRequest(githubWebBase+"/"+repo+"/releases/tag/"+url.PathEscape(tag), client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("release %s not found in %s", tag, repo)
	default:
		return fmt.Errorf("github.com returned %s for release %s", resp.Status, tag)
	}
}

// headRequest issues a HEAD request without following redirects, so the caller
// can read the Location header. GitHub answers HEAD on release URLs the same
// way as GET, minus the page body.
func headRequest(target string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "win-sshpass/"+Version)

	direct := &http.Client{
		Transport:     client.Transport,
		Timeout:       client.Timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := direct.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach GitHub: %w", err)
	}
	return resp, nil
}

func fetchReleaseFromAPI(repo, tag string, client *http.Client) (*ReleaseInfo, error) {
	if repo == "" {
		repo = DefaultUpdateRepo
	}
	if client == nil {
		client = defaultUpdateClient(updateAPITimeout)
	}

	endpoint := releaseAPIBase + "/repos/" + repo + "/releases/latest"
	if tag != "" {
		endpoint = releaseAPIBase + "/repos/" + repo + "/releases/tags/" + url.PathEscape(tag)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	// GitHub rejects API requests without a User-Agent.
	req.Header.Set("User-Agent", "win-sshpass/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusNotFound:
		if tag != "" {
			return nil, fmt.Errorf("release %s not found in %s", tag, repo)
		}
		return nil, fmt.Errorf("no releases published for %s yet", repo)
	case resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0":
		return nil, fmt.Errorf("GitHub API rate limit reached (unauthenticated requests are limited to 60 per hour); retry later or install the release manually")
	default:
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReleaseJSON))
	if err != nil {
		return nil, fmt.Errorf("cannot read the GitHub response: %w", err)
	}

	var payload struct {
		TagName     string    `json:"tag_name"`
		HTMLURL     string    `json:"html_url"`
		PublishedAt time.Time `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("cannot parse the GitHub response: %w", err)
	}
	if payload.TagName == "" {
		return nil, fmt.Errorf("the GitHub response contains no release tag")
	}

	info := &ReleaseInfo{
		Tag:         payload.TagName,
		HTMLURL:     payload.HTMLURL,
		PublishedAt: payload.PublishedAt,
	}
	for _, a := range payload.Assets {
		info.Assets = append(info.Assets, ReleaseAsset{Name: a.Name, URL: a.BrowserDownloadURL, Size: a.Size})
	}
	return info, nil
}

// SelectReleaseAsset picks the archive matching goos/goarch from a release's
// assets. Release archives are named win-sshpass-<tag>-<arch>.zip on Windows
// and win-sshpass-<tag>-<goos>-<arch>.tar.gz on Linux and macOS.
//
// The Windows .msi and macOS .pkg installers are never selected: they are
// applied by the OS package manager, not by replacing a running executable.
func SelectReleaseAsset(assets []ReleaseAsset, goos, goarch string) (*ReleaseAsset, error) {
	suffixes := assetSuffixes(goos, goarch)
	if len(suffixes) == 0 {
		return nil, fmt.Errorf("self-update is not supported on %s/%s; build from source or use a package manager", goos, goarch)
	}
	for _, suffix := range suffixes {
		for i := range assets {
			if strings.HasSuffix(assets[i].Name, suffix) {
				return &assets[i], nil
			}
		}
	}

	names := make([]string, 0, len(assets))
	for _, a := range assets {
		names = append(names, a.Name)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("the release has no downloadable files")
	}
	return nil, fmt.Errorf("the release has no build for %s/%s (looked for *%s); available: %s",
		goos, goarch, strings.Join(suffixes, " or *"), strings.Join(names, ", "))
}

// assetSuffixes returns the file-name suffixes, in order of preference, that
// identify the release archive for a platform.
func assetSuffixes(goos, goarch string) []string {
	switch goos {
	case "windows":
		return []string{"-" + goarch + ".zip"}
	case "linux", "darwin":
		return []string{"-" + goos + "-" + goarch + ".tar.gz"}
	default:
		return nil
	}
}

// archiveExt returns the archive extension of a release asset name (".zip" or
// ".tar.gz"), or "" when the asset is not an archive we can unpack.
func archiveExt(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return ".zip"
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return ".tar.gz"
	default:
		return ""
	}
}

// DownloadFile streams url into destPath. description labels the transfer for
// the progress callback; progress may be nil. client may be nil, in which case
// a default client is used.
//
// There is deliberately no ceiling on the total transfer time: release
// downloads from GitHub can be throttled to a few kilobytes per second, and a
// deadline would fail a download that is progressing perfectly well (measured:
// ~9 minutes for 3.4 MiB on a 10 KB/s link). Transfers are bounded by
// downloadStallTimeout instead, which only tolerates silence, not slowness.
func DownloadFile(url, destPath, description string, client *http.Client, progress ProgressFunc) (int64, error) {
	if client == nil {
		client = defaultUpdateClient(0)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "win-sshpass/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download failed: %s", resp.Status)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return 0, fmt.Errorf("cannot write %s: %w", destPath, err)
	}

	// Report the total up front so a progress bar can size itself before the
	// first chunk arrives; ContentLength is -1 when the server streams the
	// response without a length.
	total := resp.ContentLength
	if total < 0 {
		total = 0
	}
	if progress != nil {
		progress(description, 0, total)
	}

	// Watchdog: a connection that stops delivering bytes must not block the
	// read below forever, so closing the body from a timer unblocks it. The
	// timer is pushed back after every chunk, which is what distinguishes a
	// slow-but-alive download from a dead one.
	aborted := make(chan struct{})
	var abortOnce sync.Once
	watchdog := time.AfterFunc(downloadStallTimeout, func() {
		abortOnce.Do(func() { close(aborted) })
		resp.Body.Close()
	})
	defer watchdog.Stop()

	var written int64
	buf := make([]byte, 64<<10)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				f.Close()
				return written, fmt.Errorf("cannot write %s: %w", destPath, writeErr)
			}
			written += int64(n)
			watchdog.Reset(downloadStallTimeout)
			if progress != nil {
				progress(description, written, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			f.Close()
			select {
			case <-aborted:
				return written, fmt.Errorf("download stalled: no data for %s (received %s of %s)",
					downloadStallTimeout, humanBytes(written), humanBytes(total))
			default:
				return written, fmt.Errorf("download interrupted after %s: %w", humanBytes(written), readErr)
			}
		}
	}

	if err := f.Close(); err != nil {
		return written, fmt.Errorf("cannot write %s: %w", destPath, err)
	}
	return written, nil
}

// ExtractBinary unpacks the win-sshpass executable from a release archive
// (.zip or .tar.gz) into destPath. The executable is located by name inside
// the archive, so archives that also ship README files work unchanged.
func ExtractBinary(archivePath, destPath string) error {
	switch archiveExt(archivePath) {
	case ".zip":
		return extractFromZip(archivePath, destPath)
	case ".tar.gz":
		return extractFromTarGz(archivePath, destPath)
	default:
		return fmt.Errorf("unsupported archive format: %s", filepath.Base(archivePath))
	}
}

func extractFromZip(archivePath, destPath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", filepath.Base(archivePath), err)
	}
	defer zr.Close()

	for _, entry := range zr.File {
		if entry.FileInfo().IsDir() || !isExecutableEntry(entry.Name) {
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("cannot read %s from the archive: %w", entry.Name, err)
		}
		defer rc.Close()
		return writeExecutable(rc, destPath)
	}
	return fmt.Errorf("the archive does not contain %s", executableName())
}

func extractFromTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", filepath.Base(archivePath), err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filepath.Base(archivePath), err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", filepath.Base(archivePath), err)
		}
		if hdr.Typeflag != tar.TypeReg || !isExecutableEntry(hdr.Name) {
			continue
		}
		return writeExecutable(tr, destPath)
	}
	return fmt.Errorf("the archive does not contain %s", executableName())
}

// writeExecutable copies r into destPath as a 0755 file, refusing to grow past
// maxBinarySize so a malicious archive cannot fill the disk.
func writeExecutable(r io.Reader, destPath string) error {
	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("cannot write %s: %w", destPath, err)
	}
	defer out.Close()

	written, err := io.Copy(out, io.LimitReader(r, maxBinarySize+1))
	if err != nil {
		return fmt.Errorf("cannot unpack the executable: %w", err)
	}
	if written > maxBinarySize {
		return fmt.Errorf("the executable in the archive is larger than %s", humanBytes(maxBinarySize))
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("cannot write %s: %w", destPath, err)
	}
	return nil
}

// isExecutableEntry reports whether an archive entry is the win-sshpass
// executable, regardless of the directory it sits in.
func isExecutableEntry(name string) bool {
	base := filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	return strings.EqualFold(base, "win-sshpass") || strings.EqualFold(base, "win-sshpass.exe")
}

// executableName is the file name of the executable inside release archives.
func executableName() string {
	if runtime.GOOS == "windows" {
		return "win-sshpass.exe"
	}
	return "win-sshpass"
}

// verifyBinary rejects an extracted file that is too small or has no known
// executable header, so a failed or intercepted download cannot leave the user
// with a binary that no longer runs.
func verifyBinary(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() < minBinarySize {
		return fmt.Errorf("the downloaded file is only %s — refusing to install it over a working binary", humanBytes(fi.Size()))
	}

	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return fmt.Errorf("cannot read the downloaded file: %w", err)
	}
	if !hasExecutableMagic(magic) {
		return fmt.Errorf("the downloaded file is not an executable (header % x)", magic)
	}
	return nil
}

// hasExecutableMagic reports whether b starts with a known executable header:
// PE ("MZ") on Windows, ELF on Linux, or one of the Mach-O variants on macOS.
func hasExecutableMagic(b [4]byte) bool {
	switch {
	case b[0] == 'M' && b[1] == 'Z': // PE / DOS
		return true
	case b[0] == 0x7f && b[1] == 'E' && b[2] == 'L' && b[3] == 'F': // ELF
		return true
	case b[0] == 0xfe && b[1] == 0xed && b[2] == 0xfa: // Mach-O, big-endian
		return true
	case b[0] == 0xcf && b[1] == 0xfa && b[2] == 0xed && b[3] == 0xfe: // Mach-O 64-bit
		return true
	case b[0] == 0xce && b[1] == 0xfa && b[2] == 0xed && b[3] == 0xfe: // Mach-O 32-bit
		return true
	case b[0] == 0xca && b[1] == 0xfe && b[2] == 0xba && b[3] == 0xbe: // Mach-O universal
		return true
	}
	return false
}

// defaultUpdateClient returns the HTTP client used when the caller does not
// supply one. Cloning the default transport keeps HTTP(S)_PROXY and NO_PROXY
// support. timeout is the overall request deadline; 0 means none, which is what
// archive downloads use so that a slow link is not mistaken for a dead one.
func defaultUpdateClient(timeout time.Duration) *http.Client {
	// A host application is free to replace http.DefaultTransport, so this must
	// not be a bare type assertion: an SDK that panics in that situation would
	// take the whole process down.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Client{Timeout: timeout}
	}
	cloned := transport.Clone()
	cloned.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{Timeout: timeout, Transport: cloned}
}

// privilegedInstallHint adds a hint to a permission error when the target
// lives somewhere that normally requires elevation, which usually means the
// binary was installed by a package manager rather than by hand.
func privilegedInstallHint(target string) string {
	dir := strings.ToLower(filepath.Dir(target))
	if strings.Contains(dir, "program files") ||
		strings.Contains(dir, `\windows\system32`) ||
		strings.HasPrefix(dir, "/usr/bin") ||
		strings.HasPrefix(dir, "/usr/local/bin") {
		return "; this copy looks like it was installed by a package manager or needs administrator rights — update it with scoop/winget/MSI or re-run with elevated permissions"
	}
	return ""
}

// humanBytes formats a byte count for human-readable messages (e.g. "9.8 MiB").
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value := float64(n)
	for _, suffix := range []string{"KiB", "MiB", "GiB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TiB", value/unit)
}
