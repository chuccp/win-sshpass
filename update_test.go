package sshpass

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"1.0.0", "v1.0.0", 0},   // leading v is optional
		{"v0.9.1", "v1.0.0", -1}, // older
		{"v1.0.1", "v1.0.0", 1},  // newer
		{"v1.10.0", "v1.9.0", 1}, // numeric, not lexical
		{"v1.0", "v1.0.0", 0},    // missing parts are zero
		{"v2.0.0", "v10.0.0", -1},
		{"v1.0.0-rc1", "v1.0.0", -1}, // pre-release sorts first
		{"v1.0.0", "v1.0.0-rc1", 1},  //
		{"v1.0.0-rc1", "v1.0.0-rc2", -1},
		{"dev", "v1.0.0", -1}, // development builds are upgraded to the release
		{"v1.0.0", "dev", 1},
		{"dev", "dev", 0},
		{"", "v1.0.0", -1},
	}
	for _, tc := range cases {
		if got := CompareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSelectReleaseAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "win-sshpass-v1.0.0-amd64.zip", URL: "https://example.test/win-amd64.zip"},
		{Name: "win-sshpass-v1.0.0-arm64.zip", URL: "https://example.test/win-arm64.zip"},
		{Name: "win-sshpass-v1.0.0-amd64.msi", URL: "https://example.test/win-amd64.msi"},
		{Name: "win-sshpass-v1.0.0-linux-amd64.tar.gz", URL: "https://example.test/linux-amd64.tar.gz"},
		{Name: "win-sshpass-v1.0.0-darwin-arm64.tar.gz", URL: "https://example.test/darwin-arm64.tar.gz"},
	}

	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"windows", "amd64", "win-sshpass-v1.0.0-amd64.zip"},
		{"windows", "arm64", "win-sshpass-v1.0.0-arm64.zip"},
		{"linux", "amd64", "win-sshpass-v1.0.0-linux-amd64.tar.gz"},
		{"darwin", "arm64", "win-sshpass-v1.0.0-darwin-arm64.tar.gz"},
	}
	for _, tc := range cases {
		got, err := SelectReleaseAsset(assets, tc.goos, tc.goarch)
		if err != nil {
			t.Errorf("SelectReleaseAsset(%s/%s) failed: %v", tc.goos, tc.goarch, err)
			continue
		}
		if got.Name != tc.want {
			t.Errorf("SelectReleaseAsset(%s/%s) = %s, want %s", tc.goos, tc.goarch, got.Name, tc.want)
		}
	}

	// The .msi and .pkg installers must never be picked: they cannot be applied
	// by replacing a running executable.
	if got, err := SelectReleaseAsset(assets, "darwin", "amd64"); err == nil {
		t.Errorf("SelectReleaseAsset(darwin/amd64) = %v, want an error", got)
	}
	if _, err := SelectReleaseAsset(assets, "plan9", "amd64"); err == nil {
		t.Error("SelectReleaseAsset(plan9/amd64) succeeded, want an unsupported-platform error")
	}
	if _, err := SelectReleaseAsset(nil, "linux", "amd64"); err == nil {
		t.Error("SelectReleaseAsset(no assets) succeeded, want an error")
	}
}

func TestArchiveExt(t *testing.T) {
	cases := map[string]string{
		"win-sshpass-v1.0.0-amd64.zip":              ".zip",
		"win-sshpass-v1.0.0-amd64.msi":              "",
		"win-sshpass-v1.0.0-linux-amd64.tar.gz":     ".tar.gz",
		"win-sshpass-v1.0.0-darwin-arm64.pkg":       "",
		"win-sshpass-v1.0.0-darwin-arm64.tar.gz":    ".tar.gz",
		"/tmp/dir/with.dots/win-sshpass-v1.tgz":     ".tar.gz",
		"WIN-SSHPASS-v1.0.0-amd64.ZIP":              ".zip",
		"win-sshpass-v1.0.0-amd64.tar.gz.sha256":    "",
		"win-sshpass-v1.0.0-no-extension":           "",
		"win-sshpass-v1.0.0-linux-amd64.tarnotgz":   "",
		"win-sshpass-v1.0.0-linux-amd64.tar.gz.asc": "",
	}
	for name, want := range cases {
		if got := archiveExt(name); got != want {
			t.Errorf("archiveExt(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestIsExecutableEntry(t *testing.T) {
	yes := []string{"win-sshpass", "win-sshpass.exe", "WIN-SSHPASS.EXE", "dir/win-sshpass.exe", `dir\win-sshpass.exe`, "./win-sshpass"}
	no := []string{"README.md", "win-sshpass.txt", "win-sshpass.exe.old", "win-sshpass-v1.0.0-amd64.zip", ""}
	for _, name := range yes {
		if !isExecutableEntry(name) {
			t.Errorf("isExecutableEntry(%q) = false, want true", name)
		}
	}
	for _, name := range no {
		if isExecutableEntry(name) {
			t.Errorf("isExecutableEntry(%q) = true, want false", name)
		}
	}
}

// fakeBinary returns a payload that passes verifyBinary: a known executable
// header plus enough padding to clear the minimum-size check.
func fakeBinary(marker string) []byte {
	buf := bytes.NewBufferString("MZ" + marker)
	buf.Write(bytes.Repeat([]byte{0}, minBinarySize))
	return buf.Bytes()
}

func writeZipFixture(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "release.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTarGzFixture(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "release.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, content := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractBinary(t *testing.T) {
	want := fakeBinary("archive-payload")

	for _, tc := range []struct {
		name    string
		archive string
	}{
		{"zip", writeZipFixture(t, map[string][]byte{
			"README.md":       []byte("# win-sshpass"),
			"win-sshpass.exe": want,
			"README.zh-CN.md": []byte("# win-sshpass"),
		})},
		{"tar.gz", writeTarGzFixture(t, map[string][]byte{
			"README.md":       []byte("# win-sshpass"),
			"win-sshpass":     want,
			"README.zh-CN.md": []byte("# win-sshpass"),
		})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "extracted")
			if err := ExtractBinary(tc.archive, dest); err != nil {
				t.Fatalf("ExtractBinary failed: %v", err)
			}
			got, err := os.ReadFile(dest)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("extracted %d bytes, want %d", len(got), len(want))
			}
		})
	}
}

func TestExtractBinaryErrors(t *testing.T) {
	// an archive without the executable
	archive := writeZipFixture(t, map[string][]byte{"README.md": []byte("docs")})
	if err := ExtractBinary(archive, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("ExtractBinary succeeded for an archive without win-sshpass, want an error")
	}

	// an unsupported container
	other := filepath.Join(t.TempDir(), "release.7z")
	if err := os.WriteFile(other, []byte("not an archive"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ExtractBinary(other, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("ExtractBinary succeeded for a .7z file, want an unsupported-format error")
	}

	// a truncated zip
	broken := filepath.Join(t.TempDir(), "broken.zip")
	if err := os.WriteFile(broken, []byte("PK\x03\x04truncated"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ExtractBinary(broken, filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("ExtractBinary succeeded for a truncated zip, want an error")
	}
}

// TestReplaceExecutable covers the platform-specific swap: an atomic rename on
// Unix, rename-aside-and-replace on Windows. It runs whichever implementation
// the host compiles, so the Unix path is checked wherever the suite runs there.
func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "win-sshpass")
	staged := filepath.Join(dir, "staged")

	if err := os.WriteFile(target, []byte("old build"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new build"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Chmod explicitly: WriteFile is subject to the umask.
	if err := os.Chmod(target, 0o640); err != nil {
		t.Fatal(err)
	}

	if err := replaceExecutable(target, staged); err != nil {
		t.Fatalf("replaceExecutable failed: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new build" {
		t.Errorf("target contains %q, want the new build", got)
	}
	if _, err := os.Stat(staged); err == nil {
		t.Error("the staged file still exists after the swap")
	}
	// Nothing is running from this copy, so the parked binary must be gone on
	// every platform.
	if _, err := os.Stat(target + ".old"); err == nil {
		t.Error("the previous build was left behind as .old")
	}
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm() != 0o640 {
			t.Errorf("mode = %v, want the replaced file's 0640", fi.Mode().Perm())
		}
	}
}

func TestVerifyBinary(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good")
	if err := os.WriteFile(good, fakeBinary("ok"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := verifyBinary(good); err != nil {
		t.Errorf("verifyBinary rejected a valid binary: %v", err)
	}

	// An HTML error page saved by a captive portal must not replace the binary.
	html := filepath.Join(dir, "page")
	page := append([]byte("<!DOCTYPE html><html>"), bytes.Repeat([]byte(" "), minBinarySize)...)
	if err := os.WriteFile(html, page, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyBinary(html); err == nil {
		t.Error("verifyBinary accepted an HTML page, want an error")
	}

	// Too small to be the real binary.
	small := filepath.Join(dir, "small")
	if err := os.WriteFile(small, []byte("MZ"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := verifyBinary(small); err == nil {
		t.Error("verifyBinary accepted a 2-byte file, want an error")
	}
}

func TestDownloadFileReportsProgress(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), 200<<10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "204800")
		w.Write(payload)
	}))
	defer srv.Close()

	var calls int
	var last, total int64
	dest := filepath.Join(t.TempDir(), "downloaded")
	n, err := DownloadFile(srv.URL, dest, "Downloading test", srv.Client(), func(_ string, sent, size int64) {
		calls++
		last, total = sent, size
	})
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	if n != int64(len(payload)) {
		t.Errorf("DownloadFile wrote %d bytes, want %d", n, len(payload))
	}
	if last != int64(len(payload)) || total != int64(len(payload)) {
		t.Errorf("final progress = %d/%d, want %d/%d", last, total, len(payload), len(payload))
	}
	if calls < 2 {
		t.Errorf("progress called %d times, want at least 2 (initial + chunks)", calls)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Error("downloaded file does not match the payload")
	}
}

func TestDownloadFileAbortsOnStall(t *testing.T) {
	// Release downloads from GitHub can be slow, so the transfer must survive a
	// slow link but still give up on a connection that has gone silent.
	previous := downloadStallTimeout
	downloadStallTimeout = 200 * time.Millisecond
	defer func() { downloadStallTimeout = previous }()

	hold := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		w.Write([]byte("first chunk"))
		w.(http.Flusher).Flush()
		<-hold // stay connected but send nothing more
	}))
	defer srv.Close()
	defer close(hold)

	start := time.Now()
	_, err := DownloadFile(srv.URL, filepath.Join(t.TempDir(), "f"), "Downloading", srv.Client(), nil)
	if err == nil {
		t.Fatal("DownloadFile succeeded although the transfer stalled, want an error")
	}
	if !strings.Contains(err.Error(), "stalled") {
		t.Errorf("error = %v, want a stall error", err)
	}
	if elapsed := time.Since(start); elapsed > 30*time.Second {
		t.Errorf("stalled download took %s to abort, want it to give up promptly", elapsed)
	}
}

func TestDownloadFileErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := DownloadFile(srv.URL, filepath.Join(t.TempDir(), "f"), "Downloading", srv.Client(), nil); err == nil {
		t.Error("DownloadFile succeeded for HTTP 404, want an error")
	}
}

// fakeRelease serves a GitHub-shaped release API plus its archive downloads, so
// SelfUpdate can be exercised end to end without touching the network.
type fakeRelease struct {
	server        *httptest.Server
	tag           string
	assetNames    []string
	installers    []string
	bodies        map[string][]byte
	downloadHits  int
	downloadPaths []string // request paths the archives were fetched from
	tagHits       int      // github.com/releases/tag/<tag> lookups
	apiHits       int      // api.github.com/releases/latest lookups
	tagAPIHits    int      // api.github.com/releases/tags/<tag> lookups
}

func newFakeRelease(t *testing.T, tag string, payload []byte) *fakeRelease {
	t.Helper()

	fr := &fakeRelease{
		tag:    tag,
		bodies: map[string][]byte{},
	}

	// Publish a build for every supported platform, in the layout the release
	// workflow produces.
	archives, installers := assetNames(tag)
	fr.installers = installers
	zipBody := buildZip(t, map[string][]byte{"win-sshpass.exe": payload, "README.md": []byte("docs")})
	tarBody := buildTarGz(t, map[string][]byte{"win-sshpass": payload, "README.md": []byte("docs")})

	for _, name := range archives {
		if archiveExt(name) == ".zip" {
			fr.bodies[name] = zipBody
		} else {
			fr.bodies[name] = tarBody
		}
		fr.assetNames = append(fr.assetNames, name)
	}

	// Built inside the handler: the download URLs need the server address,
	// which only exists after httptest.NewServer returns.
	releaseJSON := func(w http.ResponseWriter) {
		assets := make([]map[string]any, 0, len(fr.assetNames)+len(fr.installers))
		// The .msi/.pkg installers are part of the release too; the updater has
		// to skip them, because they are applied by the OS package manager
		// rather than by replacing a running executable.
		for _, name := range append(append([]string{}, fr.assetNames...), fr.installers...) {
			assets = append(assets, map[string]any{
				"name":                 name,
				"browser_download_url": fr.server.URL + "/dl/" + name,
				"size":                 1024,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"tag_name": tag,
			"html_url": "https://github.com/" + DefaultUpdateRepo + "/releases/tag/" + tag,
			"assets":   assets,
		})
	}

	// Download handler shared by both routes: the API's browser_download_url
	// (/dl/<name>) and github.com's tag-pinned /releases/download/<tag>/<name>.
	serve := func(w http.ResponseWriter, r *http.Request) {
		body, ok := fr.bodies[path.Base(r.URL.Path)]
		if !ok {
			// An installer or an unknown asset: nothing to serve, and the
			// updater must never ask for one in the first place.
			http.NotFound(w, r)
			return
		}
		fr.downloadHits++
		fr.downloadPaths = append(fr.downloadPaths, r.URL.Path)
		w.Write(body)
	}

	mux := http.NewServeMux()
	// --- github.com routes (the primary lookup) ---
	mux.HandleFunc("/"+DefaultUpdateRepo+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/"+DefaultUpdateRepo+"/releases/tag/"+tag, http.StatusFound)
	})
	mux.HandleFunc("/"+DefaultUpdateRepo+"/releases/tag/", func(w http.ResponseWriter, r *http.Request) {
		if path.Base(r.URL.Path) != tag {
			http.NotFound(w, r)
			return
		}
		fr.tagHits++
		w.WriteHeader(http.StatusOK)
	})
	// Only the tag-pinned download route is served. /releases/latest/download/
	// is deliberately left unrouteable: it resolves to whatever is newest when
	// the download starts, so a release published between the version check and
	// the download would serve a different build (or 404). If the updater ever
	// goes back to it, these tests fail with a 404 instead of passing quietly.
	mux.HandleFunc("/"+DefaultUpdateRepo+"/releases/download/", serve)

	// --- api.github.com routes (the fallback) ---
	mux.HandleFunc("/repos/"+DefaultUpdateRepo+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fr.apiHits++
		releaseJSON(w)
	})
	mux.HandleFunc("/repos/"+DefaultUpdateRepo+"/releases/tags/"+tag, func(w http.ResponseWriter, r *http.Request) {
		fr.tagAPIHits++
		releaseJSON(w)
	})
	mux.HandleFunc("/dl/", serve)
	fr.server = httptest.NewServer(mux)
	t.Cleanup(fr.server.Close)

	return fr
}

// useReleaseAPI points both release routes (github.com and api.github.com) at
// a fake server, so no test ever touches the network.
func useReleaseAPI(t *testing.T, fr *fakeRelease) {
	t.Helper()
	previousAPI, previousWeb := releaseAPIBase, githubWebBase
	releaseAPIBase, githubWebBase = fr.server.URL, fr.server.URL
	t.Cleanup(func() {
		releaseAPIBase, githubWebBase = previousAPI, previousWeb
	})
}

// deadServer returns the URL of a server that has already been shut down, so
// requests to it fail immediately with a connection refused.
func deadServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	return url
}

func testTargetPath(t *testing.T) string {
	t.Helper()
	name := "win-sshpass"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(t.TempDir(), name)
}

func TestSelfUpdateInstallsNewRelease(t *testing.T) {
	const tag = "v99.99.99" // newer than any real version
	payload := fakeBinary("release-payload")
	fr := newFakeRelease(t, tag, payload)
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if !res.UpdateAvailable || !res.Updated {
		t.Fatalf("SelfUpdate = %+v, want an available and applied update", res)
	}
	if res.CurrentVersion != Version || res.LatestVersion != tag {
		t.Errorf("versions = %s -> %s, want %s -> %s", res.CurrentVersion, res.LatestVersion, Version, tag)
	}
	if res.AssetName == "" || archiveExt(res.AssetName) == "" {
		t.Errorf("asset %q does not look like a release archive", res.AssetName)
	}
	if res.ReleaseURL == "" {
		t.Error("UpdateResult.ReleaseURL is empty")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("target holds %d bytes, want the %d-byte release binary", len(got), len(payload))
	}
	if _, err := os.Stat(target + ".old"); err == nil {
		t.Error("the previous binary was left behind as .old")
	}
}

// The update must not depend on api.github.com: it is rate limited to 60
// unauthenticated requests per hour and is unreachable on some networks where
// github.com works fine. The normal route is github.com's /releases/latest
// redirect plus a download URL derived from the tag.
func TestSelfUpdateUsesWebRouteWithoutAPI(t *testing.T) {
	const tag = "v99.99.97"
	payload := fakeBinary("web-route")
	fr := newFakeRelease(t, tag, payload)
	useReleaseAPI(t, fr)
	releaseAPIBase = deadServer(t)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if !res.Updated || res.LatestVersion != tag {
		t.Fatalf("SelfUpdate = %+v, want %s installed", res, tag)
	}
	if fr.apiHits != 0 {
		t.Errorf("the GitHub API was queried %d times, want 0", fr.apiHits)
	}
	// The archive must come from the immutable, tag-pinned URL rather than
	// /releases/latest/download/, which can hand back a different release.
	wantPath := "/" + DefaultUpdateRepo + "/releases/download/" + tag + "/"
	if len(fr.downloadPaths) != 1 || !strings.HasPrefix(fr.downloadPaths[0], wantPath) {
		t.Errorf("downloaded from %v, want a single %s<asset> request", fr.downloadPaths, wantPath)
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, payload) {
		t.Error("target does not hold the release binary")
	}
}

// When github.com is unreachable the API route must still carry the update.
func TestSelfUpdateFallsBackToAPI(t *testing.T) {
	const tag = "v99.99.96"
	payload := fakeBinary("api-route")
	fr := newFakeRelease(t, tag, payload)
	useReleaseAPI(t, fr)
	githubWebBase = deadServer(t)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if !res.Updated || res.LatestVersion != tag {
		t.Fatalf("SelfUpdate = %+v, want %s installed", res, tag)
	}
	if fr.apiHits != 1 {
		t.Errorf("GitHub API hits = %d, want 1", fr.apiHits)
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, payload) {
		t.Error("target does not hold the release binary")
	}
}

// A pinned tag that does not exist must be rejected before anything is
// downloaded.
func TestSelfUpdateUnknownTag(t *testing.T) {
	fr := newFakeRelease(t, "v99.99.95", fakeBinary("x"))
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := SelfUpdate(UpdateOptions{TargetPath: target, Version: "v0.0.1"}); err == nil {
		t.Fatal("SelfUpdate succeeded for an unknown tag, want an error")
	}
	if fr.downloadHits != 0 {
		t.Error("an archive was downloaded for an unknown tag")
	}
	if got, _ := os.ReadFile(target); string(got) != "previous build" {
		t.Error("target was modified")
	}
}

func TestSelfUpdateSkipsWhenUpToDate(t *testing.T) {
	fr := newFakeRelease(t, Version, fakeBinary("same"))
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if res.UpdateAvailable || res.Updated {
		t.Errorf("SelfUpdate = %+v, want no update for the running version", res)
	}
	if fr.downloadHits != 0 {
		t.Errorf("downloaded %d archives although the build is up to date", fr.downloadHits)
	}
	if got, _ := os.ReadFile(target); string(got) != "previous build" {
		t.Error("target was modified although the build is up to date")
	}
}

func TestSelfUpdateCheckOnly(t *testing.T) {
	const tag = "v99.99.99"
	fr := newFakeRelease(t, tag, fakeBinary("release-payload"))
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target, CheckOnly: true})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if !res.UpdateAvailable {
		t.Error("UpdateAvailable = false, want true")
	}
	if res.Updated {
		t.Error("Updated = true, want false for a check-only run")
	}
	if fr.downloadHits != 0 {
		t.Errorf("downloaded %d archives in check-only mode", fr.downloadHits)
	}
	if got, _ := os.ReadFile(target); string(got) != "previous build" {
		t.Error("target was modified in check-only mode")
	}
}

func TestSelfUpdateForceReinstalls(t *testing.T) {
	payload := fakeBinary("release-payload")
	fr := newFakeRelease(t, Version, payload)
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target, Force: true})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if res.UpdateAvailable {
		t.Error("UpdateAvailable = true, want false when reinstalling the same version")
	}
	if !res.Updated {
		t.Error("Updated = false, want true for a forced reinstall")
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, payload) {
		t.Error("target was not replaced by the forced reinstall")
	}
}

func TestSelfUpdateByTag(t *testing.T) {
	const tag = "v99.99.98"
	payload := fakeBinary("pinned")
	fr := newFakeRelease(t, tag, payload)
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := SelfUpdate(UpdateOptions{TargetPath: target, Version: tag})
	if err != nil {
		t.Fatalf("SelfUpdate failed: %v", err)
	}
	if fr.tagHits != 1 {
		t.Errorf("github.com tag lookup hit %d times, want 1", fr.tagHits)
	}
	if fr.tagAPIHits != 0 {
		t.Errorf("the GitHub API was queried %d times for a pinned tag, want 0", fr.tagAPIHits)
	}
	if !res.Updated || res.LatestVersion != tag {
		t.Errorf("SelfUpdate = %+v, want %s installed", res, tag)
	}
}

func TestSelfUpdateRejectsNonExecutableDownload(t *testing.T) {
	const tag = "v99.99.99"
	// A valid archive that does not contain a usable binary — the shape of a
	// captive portal or proxy returning an HTML page instead of the release.
	fr := newFakeRelease(t, tag, []byte("<html><body>blocked</body></html>"))
	useReleaseAPI(t, fr)

	target := testTargetPath(t)
	if err := os.WriteFile(target, []byte("previous build"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := SelfUpdate(UpdateOptions{TargetPath: target}); err == nil {
		t.Fatal("SelfUpdate succeeded for a non-executable download, want an error")
	}
	if got, _ := os.ReadFile(target); string(got) != "previous build" {
		t.Error("target was replaced by a non-executable download")
	}
}

func TestSelfUpdateUnwritableTarget(t *testing.T) {
	const tag = "v99.99.99"
	fr := newFakeRelease(t, tag, fakeBinary("release-payload"))
	useReleaseAPI(t, fr)

	// The staging file is created next to the target, so a missing directory
	// must fail before anything is replaced.
	missing := filepath.Join(t.TempDir(), "does-not-exist", "win-sshpass")
	if _, err := SelfUpdate(UpdateOptions{TargetPath: missing}); err == nil {
		t.Fatal("SelfUpdate succeeded for an unwritable target directory, want an error")
	}
	if fr.downloadHits != 0 {
		t.Error("the archive was downloaded although the target directory is unusable")
	}
}

func TestFetchLatestReleaseNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	previous := releaseAPIBase
	releaseAPIBase = srv.URL
	defer func() { releaseAPIBase = previous }()

	if _, err := FetchLatestRelease(DefaultUpdateRepo, srv.Client()); err == nil {
		t.Error("FetchLatestRelease succeeded for a missing repository, want an error")
	}
}

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		0:                "0 B",
		512:              "512 B",
		2048:             "2.0 KiB",
		10 << 20:         "10.0 MiB",
		3 << 30:          "3.0 GiB",
		5 << 40:          "5.0 TiB",
		1536:             "1.5 KiB",
		1024 * 1024 * 26: "26.0 MiB",
	}
	for n, want := range cases {
		if got := humanBytes(n); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", n, got, want)
		}
	}
}

func buildZip(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildTarGz(t *testing.T, entries map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// assetNames lists the files attached to a release: the archives the updater
// may install, and the installers it must leave alone.
func assetNames(tag string) (archives, installers []string) {
	for _, arch := range []string{"amd64", "arm64"} {
		archives = append(archives,
			"win-sshpass-"+tag+"-"+arch+".zip",
			"win-sshpass-"+tag+"-linux-"+arch+".tar.gz",
			"win-sshpass-"+tag+"-darwin-"+arch+".tar.gz")
		installers = append(installers,
			"win-sshpass-"+tag+"-"+arch+".msi",
			"win-sshpass-"+tag+"-darwin-"+arch+".pkg")
	}
	return archives, installers
}
