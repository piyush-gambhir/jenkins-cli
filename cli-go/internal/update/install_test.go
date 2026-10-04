package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

type archiveEntry struct {
	name     string
	body     string
	typeflag byte   // tar only; 0 means a regular file
	mode     uint32 // zip only; 0 means a regular file
}

func makeTarGz(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: tar.TypeReg}
		if e.typeflag != 0 && e.typeflag != tar.TypeReg {
			hdr.Typeflag, hdr.Size, hdr.Linkname = e.typeflag, 0, "/etc/passwd"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Size > 0 {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
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

func makeZip(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		fh := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		mode := os.FileMode(0o755)
		if e.mode != 0 {
			mode = os.FileMode(e.mode)
		}
		fh.SetMode(mode)
		w, err := zw.CreateHeader(fh)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// releaseServer serves one release's archive and checksums.txt and counts
// requests.
func releaseServer(t *testing.T, version, archiveName string, archive []byte, checksums string) (string, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	prefix := "/" + Repo + "/releases/download/v" + version + "/"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case prefix + "checksums.txt":
			fmt.Fprint(w, checksums)
		case prefix + archiveName:
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL, &hits
}

// setupInstall writes an "old" executable and returns an installer for goos
// pointed at a release server.
func setupInstall(t *testing.T, goos string, archive []byte, checksums func(name string, archive []byte) string) (*Installer, *atomic.Int32) {
	t.Helper()
	in := &Installer{GOOS: goos, GOARCH: "amd64", Rename: os.Rename, Client: http.DefaultClient}
	exe := filepath.Join(t.TempDir(), in.BinaryName())
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	in.ExecPath = exe
	name := in.ArchiveName()
	base, hits := releaseServer(t, "0.2.9", name, archive, checksums(name, archive))
	in.BaseURL = base
	return in, hits
}

func goodSums(name string, archive []byte) string {
	return fmt.Sprintf("%s  other_file.tar.gz\n%s  %s\n", strings.Repeat("0", 64), sha(archive), name)
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s contains %q, want %q", path, got, want)
	}
}

func TestArchiveNamesMatchGoReleaser(t *testing.T) {
	for _, tc := range []struct{ goos, goarch, archive, binary string }{
		{"darwin", "arm64", "jenkins-cli_darwin_arm64.tar.gz", "jenkins"},
		{"linux", "amd64", "jenkins-cli_linux_amd64.tar.gz", "jenkins"},
		{"windows", "amd64", "jenkins-cli_windows_amd64.zip", "jenkins.exe"},
	} {
		in := &Installer{GOOS: tc.goos, GOARCH: tc.goarch}
		if in.ArchiveName() != tc.archive || in.BinaryName() != tc.binary {
			t.Errorf("%s/%s: got %s, %s; want %s, %s", tc.goos, tc.goarch, in.ArchiveName(), in.BinaryName(), tc.archive, tc.binary)
		}
	}
}

func TestInstallReplacesUnixBinary(t *testing.T) {
	archive := makeTarGz(t, archiveEntry{name: "LICENSE", body: "license"}, archiveEntry{name: "jenkins", body: "new binary"})
	in, _ := setupInstall(t, "linux", archive, goodSums)
	if err := in.Install(context.Background(), "0.2.9"); err != nil {
		t.Fatal(err)
	}
	assertContent(t, in.ExecPath, "new binary")
	if runtime.GOOS != "windows" {
		st, err := os.Stat(in.ExecPath)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o755 {
			t.Errorf("new binary mode %v, want 0755", st.Mode().Perm())
		}
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(in.ExecPath), ".jenkins-*"))
	if len(leftovers) != 0 {
		t.Errorf("temp files left next to the binary: %v", leftovers)
	}
}

func TestInstallWindowsMovesRunningBinaryAside(t *testing.T) {
	archive := makeZip(t, archiveEntry{name: "README.md", body: "readme"}, archiveEntry{name: "jenkins.exe", body: "new exe"})
	in, _ := setupInstall(t, "windows", archive, goodSums)
	// A leftover from an earlier update must not block this one.
	if err := os.WriteFile(in.ExecPath+".old", []byte("older exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	var renames []string
	in.Rename = func(oldpath, newpath string) error {
		renames = append(renames, filepath.Base(oldpath)+" -> "+filepath.Base(newpath))
		return os.Rename(oldpath, newpath)
	}
	if err := in.Install(context.Background(), "0.2.9"); err != nil {
		t.Fatal(err)
	}
	assertContent(t, in.ExecPath, "new exe")
	assertContent(t, in.ExecPath+".old", "old binary")
	if len(renames) != 2 || renames[0] != "jenkins.exe -> jenkins.exe.old" || !strings.HasSuffix(renames[1], " -> jenkins.exe") {
		t.Errorf("renames = %v, want the running exe moved aside before the new one moves in", renames)
	}
}

func TestInstallWindowsRestoresOldBinaryWhenMoveFails(t *testing.T) {
	archive := makeZip(t, archiveEntry{name: "jenkins.exe", body: "new exe"})
	in, _ := setupInstall(t, "windows", archive, goodSums)
	calls := 0
	in.Rename = func(oldpath, newpath string) error {
		calls++
		if calls == 2 {
			return errors.New("access denied")
		}
		return os.Rename(oldpath, newpath)
	}
	if err := in.Install(context.Background(), "0.2.9"); err == nil {
		t.Fatal("expected the failed move to be reported")
	}
	assertContent(t, in.ExecPath, "old binary")
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(in.ExecPath), ".jenkins-*"))
	if len(leftovers) != 0 {
		t.Errorf("temp files left next to the binary: %v", leftovers)
	}
}

func TestInstallRefusesChecksumMismatch(t *testing.T) {
	archive := makeTarGz(t, archiveEntry{name: "jenkins", body: "tampered"})
	in, _ := setupInstall(t, "linux", archive, func(name string, _ []byte) string {
		return fmt.Sprintf("%s  %s\n", sha([]byte("something else")), name)
	})
	err := in.Install(context.Background(), "0.2.9")
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("Install = %v, want a checksum mismatch", err)
	}
	assertContent(t, in.ExecPath, "old binary")
}

func TestInstallRefusesMissingChecksum(t *testing.T) {
	archive := makeTarGz(t, archiveEntry{name: "jenkins", body: "new binary"})
	in, hits := setupInstall(t, "linux", archive, func(string, []byte) string {
		return sha([]byte("x")) + "  jenkins-cli_plan9_amd64.tar.gz\n"
	})
	err := in.Install(context.Background(), "0.2.9")
	if err == nil || !strings.Contains(err.Error(), "no checksum for jenkins-cli_linux_amd64.tar.gz") {
		t.Fatalf("Install = %v, want a missing checksum refusal", err)
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("%d requests; the archive should not be downloaded without a checksum", n)
	}
	assertContent(t, in.ExecPath, "old binary")
}

func TestInstallRefusesUnsafeArchives(t *testing.T) {
	for _, tc := range []struct {
		name    string
		goos    string
		archive func(t *testing.T) []byte
		want    string
	}{
		{"tar traversal", "linux", func(t *testing.T) []byte {
			return makeTarGz(t, archiveEntry{name: "../../jenkins", body: "evil"}, archiveEntry{name: "jenkins", body: "new"})
		}, "unsafe path"},
		{"tar absolute path", "linux", func(t *testing.T) []byte {
			return makeTarGz(t, archiveEntry{name: "/usr/local/bin/jenkins", body: "evil"})
		}, "unsafe path"},
		{"tar symlink binary", "linux", func(t *testing.T) []byte {
			return makeTarGz(t, archiveEntry{name: "jenkins", typeflag: tar.TypeSymlink})
		}, "not a regular file"},
		{"tar hardlink binary", "linux", func(t *testing.T) []byte {
			return makeTarGz(t, archiveEntry{name: "jenkins", typeflag: tar.TypeLink})
		}, "not a regular file"},
		{"tar without binary", "linux", func(t *testing.T) []byte {
			return makeTarGz(t, archiveEntry{name: "README.md", body: "readme"})
		}, "not found"},
		{"zip traversal", "windows", func(t *testing.T) []byte {
			return makeZip(t, archiveEntry{name: "../jenkins.exe", body: "evil"})
		}, "unsafe path"},
		{"zip backslash traversal", "windows", func(t *testing.T) []byte {
			return makeZip(t, archiveEntry{name: `..\jenkins.exe`, body: "evil"})
		}, "unsafe path"},
		{"zip symlink binary", "windows", func(t *testing.T) []byte {
			return makeZip(t, archiveEntry{name: "jenkins.exe", body: "C:/Windows/System32/cmd.exe", mode: uint32(os.ModeSymlink | 0o777)})
		}, "not a regular file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, _ := setupInstall(t, tc.goos, tc.archive(t), goodSums)
			err := in.Install(context.Background(), "0.2.9")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Install = %v, want an error containing %q", err, tc.want)
			}
			assertContent(t, in.ExecPath, "old binary")
			entries, _ := os.ReadDir(filepath.Dir(in.ExecPath))
			if len(entries) != 1 {
				t.Errorf("directory has %d entries after a refused install, want only the old binary", len(entries))
			}
		})
	}
}

func TestInstallEnforcesSizeLimit(t *testing.T) {
	old := maxArtifactBytes
	maxArtifactBytes = 4096
	t.Cleanup(func() { maxArtifactBytes = old })
	// The archive is small, but the binary inside expands past the limit.
	archive := makeTarGz(t, archiveEntry{name: "jenkins", body: strings.Repeat("a", 100000)})
	if int64(len(archive)) > maxArtifactBytes {
		t.Fatalf("compressed test archive is %d bytes, over the limit", len(archive))
	}
	in, _ := setupInstall(t, "linux", archive, goodSums)
	err := in.Install(context.Background(), "0.2.9")
	if err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("Install = %v, want a size limit error", err)
	}
	assertContent(t, in.ExecPath, "old binary")
}

func TestInstallUnwritableDirectoryKeepsOldBinary(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions and a non-root user")
	}
	archive := makeTarGz(t, archiveEntry{name: "jenkins", body: "new binary"})
	in, hits := setupInstall(t, "linux", archive, goodSums)
	dir := filepath.Dir(in.ExecPath)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := in.Install(context.Background(), "0.2.9")
	var notWritable *NotWritableError
	if !errors.As(err, &notWritable) {
		t.Fatalf("Install = %v, want NotWritableError", err)
	}
	for _, want := range []string{dir, "sudo jenkins update", "install.sh", "INSTALL_DIR="} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q:\n%v", want, err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("%d downloads before discovering the directory is not writable", n)
	}
	assertContent(t, in.ExecPath, "old binary")
}
