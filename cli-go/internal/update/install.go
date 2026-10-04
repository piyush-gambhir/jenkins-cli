package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// projectName matches project_name in cli-go/.goreleaser.yaml, which
	// names the archives <project>_<os>_<arch>.tar.gz (.zip on Windows).
	projectName      = "jenkins-cli"
	installScriptURL = "https://raw.githubusercontent.com/piyush-gambhir/jenkins-cli/main/install.sh"
	maxChecksumBytes = 1 << 20
	downloadTimeout  = 120 * time.Second
)

// maxArtifactBytes caps both the downloaded archive and the extracted
// binary; tests lower it.
var maxArtifactBytes int64 = 256 << 20

// Installer downloads a release archive, verifies it against the release's
// checksums.txt, and replaces the executable at ExecPath. Any failure leaves
// the existing executable in place.
type Installer struct {
	GOOS, GOARCH string
	ExecPath     string // resolved path of the running executable
	BaseURL      string // release download host, https://github.com by default
	Client       *http.Client
	// Rename is os.Rename; tests replace it to simulate failures.
	Rename func(oldpath, newpath string) error
}

// NewInstaller returns an installer for the running platform.
func NewInstaller(execPath string) *Installer {
	return &Installer{
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,
		ExecPath: execPath,
		BaseURL:  "https://github.com",
		Client: &http.Client{
			Timeout: downloadTimeout,
			// Release assets redirect to a CDN; never follow a downgrade to http.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("too many redirects")
				}
				if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
					return fmt.Errorf("refusing redirect to non-HTTPS URL %s", req.URL.Redacted())
				}
				return nil
			},
		},
		Rename: os.Rename,
	}
}

// NotWritableError means the executable's directory cannot be written, so
// the binary cannot be replaced.
type NotWritableError struct {
	Dir  string
	GOOS string
	Err  error
}

func (e *NotWritableError) Error() string {
	cause := e.Err
	var pathErr *os.PathError
	if errors.As(e.Err, &pathErr) {
		cause = pathErr.Err
	}
	if e.GOOS == "windows" {
		return fmt.Sprintf("cannot write to %s (%v); %s was not changed.\nRe-run from a terminal opened as Administrator, or move %s.exe to a directory you can write to.",
			e.Dir, cause, Binary, Binary)
	}
	return fmt.Sprintf("cannot write to %s (%v); %s was not changed.\nRe-run with sudo: sudo %s update\nOr reinstall into a directory you can write to: curl -sSfL %s | INSTALL_DIR=~/.local/bin sh",
		e.Dir, cause, Binary, Binary, installScriptURL)
}

func (e *NotWritableError) Unwrap() error { return e.Err }

// ArchiveName is the GoReleaser archive for the installer's platform.
func (in *Installer) ArchiveName() string {
	ext := "tar.gz"
	if in.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("%s_%s_%s.%s", projectName, in.GOOS, in.GOARCH, ext)
}

// BinaryName is the executable name inside the archive.
func (in *Installer) BinaryName() string {
	if in.GOOS == "windows" {
		return Binary + ".exe"
	}
	return Binary
}

// Install downloads version, verifies its SHA-256 checksum, and swaps it in
// for the executable at ExecPath.
func (in *Installer) Install(ctx context.Context, version string) error {
	version = NormalizeVersion(version)
	if !releaseVersion.MatchString(version) {
		return fmt.Errorf("invalid release version %q", version)
	}
	dir := filepath.Dir(in.ExecPath)
	if err := checkWritable(dir); err != nil {
		return &NotWritableError{Dir: dir, GOOS: in.GOOS, Err: err}
	}

	archive := in.ArchiveName()
	base := fmt.Sprintf("%s/%s/releases/download/v%s/", strings.TrimRight(in.BaseURL, "/"), Repo, version)

	var sums bytes.Buffer
	if err := in.download(ctx, base+"checksums.txt", &sums, maxChecksumBytes); err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	want, err := checksumFor(sums.Bytes(), archive)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "jenkins-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archive)
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	hash := sha256.New()
	err = in.download(ctx, base+archive, io.MultiWriter(f, hash), maxArtifactBytes)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("downloading %s: %w", archive, err)
	}
	if got := hex.EncodeToString(hash.Sum(nil)); got != want {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s; refusing to install", archive, want, got)
	}

	newPath, err := in.extract(archivePath, dir)
	if err != nil {
		return fmt.Errorf("extracting %s: %w", archive, err)
	}
	if err := in.replace(newPath); err != nil {
		_ = os.Remove(newPath)
		return err
	}
	return nil
}

func (in *Installer) download(ctx context.Context, url string, dst io.Writer, limit int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", projectName)
	resp, err := in.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned status %d", url, resp.StatusCode)
	}
	return copyLimited(dst, resp.Body, limit)
}

// copyLimited copies src to dst and fails if src is larger than limit.
func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	n, err := io.Copy(dst, io.LimitReader(src, limit+1))
	if err != nil {
		return err
	}
	if n > limit {
		return fmt.Errorf("exceeds the %d byte size limit", limit)
	}
	return nil
}

// checksumFor finds archive's SHA-256 in a GoReleaser checksums.txt
// ("<hex>  <name>" per line).
func checksumFor(sums []byte, archive string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(sums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != archive {
			continue
		}
		sum := strings.ToLower(fields[0])
		if b, err := hex.DecodeString(sum); err != nil || len(b) != sha256.Size {
			return "", fmt.Errorf("malformed checksum for %s in checksums.txt", archive)
		}
		return sum, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading checksums.txt: %w", err)
	}
	return "", fmt.Errorf("no checksum for %s in checksums.txt; refusing to install (is there a release build for this platform?)", archive)
}

// extract writes only the expected binary from the archive into a new temp
// file in dir and returns its path. Entries with unsafe paths, and a binary
// entry that is not a regular file, make the whole archive invalid.
func (in *Installer) extract(archivePath, dir string) (string, error) {
	out, err := os.CreateTemp(dir, "."+Binary+"-update-*")
	if err != nil {
		return "", err
	}
	tmp := out.Name()
	if in.GOOS == "windows" {
		err = extractZip(archivePath, in.BinaryName(), out)
	} else {
		err = extractTarGz(archivePath, in.BinaryName(), out)
	}
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp, 0o755)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return tmp, nil
}

// entryIsBinary validates an archive entry name and reports whether it is
// the binary (at the archive root or in a subdirectory).
func entryIsBinary(name, binary string) (bool, error) {
	if name == "" || strings.Contains(name, `\`) || path.IsAbs(name) {
		return false, fmt.Errorf("unsafe path %q in archive", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return false, fmt.Errorf("unsafe path %q in archive", name)
		}
	}
	return path.Base(path.Clean(name)) == binary, nil
}

func extractTarGz(archivePath, binary string, out io.Writer) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("%s not found in archive", binary)
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
		isBinary, err := entryIsBinary(hdr.Name, binary)
		if err != nil {
			return err
		}
		if !isBinary {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			return fmt.Errorf("%s in archive is not a regular file", hdr.Name)
		}
		if hdr.Size > maxArtifactBytes {
			return fmt.Errorf("%s exceeds the %d byte size limit", hdr.Name, maxArtifactBytes)
		}
		return copyLimited(out, tr, maxArtifactBytes)
	}
}

func extractZip(archivePath, binary string, out io.Writer) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		isBinary, err := entryIsBinary(zf.Name, binary)
		if err != nil {
			return err
		}
		if !isBinary {
			continue
		}
		if !zf.Mode().IsRegular() {
			return fmt.Errorf("%s in archive is not a regular file", zf.Name)
		}
		if zf.UncompressedSize64 > uint64(maxArtifactBytes) {
			return fmt.Errorf("%s exceeds the %d byte size limit", zf.Name, maxArtifactBytes)
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		err = copyLimited(out, rc, maxArtifactBytes)
		rc.Close()
		return err
	}
	return fmt.Errorf("%s not found in archive", binary)
}

// replace moves newPath over ExecPath. On Unix a rename replaces the file
// atomically. Windows cannot overwrite a running .exe but can rename it, so
// the old binary moves aside to <exe>.old (deleted on a later start) and is
// moved back if the new one cannot be put in place.
func (in *Installer) replace(newPath string) error {
	if in.GOOS != "windows" {
		if err := in.Rename(newPath, in.ExecPath); err != nil {
			return fmt.Errorf("replacing %s: %w", in.ExecPath, err)
		}
		return nil
	}
	old := in.ExecPath + ".old"
	if err := os.Remove(old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing leftover %s: %w", old, err)
	}
	if err := in.Rename(in.ExecPath, old); err != nil {
		return fmt.Errorf("moving %s aside: %w", in.ExecPath, err)
	}
	if err := in.Rename(newPath, in.ExecPath); err != nil {
		if rerr := in.Rename(old, in.ExecPath); rerr != nil {
			return fmt.Errorf("installing new binary: %w (restoring the old one also failed: %v; it is at %s)", err, rerr, old)
		}
		return fmt.Errorf("installing new binary: %w", err)
	}
	return nil
}

func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, "."+Binary+"-write-test-*")
	if err != nil {
		return err
	}
	name := f.Name()
	f.Close()
	return os.Remove(name)
}
