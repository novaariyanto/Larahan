package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ProgressFunc receives download byte progress.
type ProgressFunc func(written, total int64)

// Client downloads remote packages into the local cache.
type Client struct {
	httpClient *http.Client
	cache      *Cache
}

// NewClient creates a downloader client.
func NewClient(cache *Cache) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 0}, // large archives; rely on context
		cache:      cache,
	}
}

// DownloadResult describes the outcome of Ensure.
type DownloadResult struct {
	LocalPath string
	Cached    bool
	Bytes     int64
}

// Ensure returns a local package path, downloading only when cache miss.
func (c *Client) Ensure(ctx context.Context, spec PackageSpec, onProgress ProgressFunc) (DownloadResult, error) {
	if spec.URL == "" {
		return DownloadResult{}, fmt.Errorf("package URL belum dikonfigurasi untuk %s@%s", spec.Type, spec.Version)
	}
	if err := c.cache.EnsureDir(); err != nil {
		return DownloadResult{}, err
	}

	path, ok, err := c.cache.HasValid(spec.Filename, spec.SizeHint)
	if err != nil {
		return DownloadResult{}, err
	}
	if ok {
		info, _ := os.Stat(path)
		var size int64
		if info != nil {
			size = info.Size()
		}
		if onProgress != nil {
			onProgress(size, size)
		}
		return DownloadResult{LocalPath: path, Cached: true, Bytes: size}, nil
	}

	return c.download(ctx, spec, path, onProgress)
}

func (c *Client) download(ctx context.Context, spec PackageSpec, dest string, onProgress ProgressFunc) (DownloadResult, error) {
	tmp := dest + ".partial"
	_ = os.Remove(tmp)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.URL, nil)
	if err != nil {
		return DownloadResult{}, err
	}
	req.Header.Set("User-Agent", "Larahan/0.1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("download %s: %w", spec.Filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DownloadResult{}, fmt.Errorf("download %s: HTTP %d", spec.Filename, resp.StatusCode)
	}

	out, err := os.Create(tmp)
	if err != nil {
		return DownloadResult{}, err
	}

	total := resp.ContentLength
	if spec.SizeHint > 0 {
		total = spec.SizeHint
	}

	writer := &progressWriter{
		w:          out,
		total:      total,
		onProgress: onProgress,
		minInterval: 200 * time.Millisecond,
	}

	written, err := io.Copy(writer, resp.Body)
	closeErr := out.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return DownloadResult{}, fmt.Errorf("write download: %w", err)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return DownloadResult{}, closeErr
	}

	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return DownloadResult{}, fmt.Errorf("finalize download: %w", err)
	}

	if onProgress != nil {
		onProgress(written, written)
	}

	return DownloadResult{LocalPath: dest, Cached: false, Bytes: written}, nil
}

type progressWriter struct {
	w           io.Writer
	written     int64
	total       int64
	onProgress  ProgressFunc
	lastEmit    time.Time
	minInterval time.Duration
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.written += int64(n)
	if p.onProgress != nil {
		now := time.Now()
		if now.Sub(p.lastEmit) >= p.minInterval || (p.total > 0 && p.written >= p.total) {
			p.onProgress(p.written, p.total)
			p.lastEmit = now
		}
	}
	return n, err
}
