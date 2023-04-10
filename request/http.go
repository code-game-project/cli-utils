package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"

	neturl "net/url"
)

var (
	httpCacheDir = filepath.Join(xdg.CacheHome, "codegame", "http")
	etagCacheDir = filepath.Join(httpCacheDir, "etag")
)

var errNoETag = errors.New("no etag")

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type cacheReader struct {
	r io.ReadCloser
	w io.WriteCloser
}

func (r *cacheReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		if n, err := r.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (r *cacheReader) Close() error {
	r.w.Close()
	return r.r.Close()
}

func FetchFile(url string, cacheMaxAge time.Duration) (io.ReadCloser, error) {
	cacheFilePath := filepath.Join(httpCacheDir, neturl.PathEscape(url))
	if cacheMaxAge > 0 {
		if stat, err := os.Stat(cacheFilePath); err == nil && time.Since(stat.ModTime()) <= cacheMaxAge {
			file, err := os.Open(cacheFilePath)
			if err == nil {
				return file, nil
			}
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	loadETag(url, req)
	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode == http.StatusNotModified {
		file, err2 := os.Open(cacheFilePath)
		if err2 == nil {
			if err == nil {
				os.Chtimes(cacheFilePath, time.Now(), time.Now())
			}
			return file, nil
		}
		os.Remove(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch '%s': status '%s'", url, resp.Status)
	}
	saveETag(url, resp)

	reader := resp.Body
	if cacheMaxAge > 0 {
		err := os.MkdirAll(httpCacheDir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
		file, err := os.Create(cacheFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create cache file: %w", err)
		}
		reader = &cacheReader{
			r: resp.Body,
			w: file,
		}
	}
	return reader, nil
}

func FetchJSON[T any](url string, maxCacheAge time.Duration) (T, error) {
	var obj T
	file, err := FetchFile(url, maxCacheAge)
	if err != nil {
		return obj, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("failed to decode response from '%s': %w", url, err)
	}
	return obj, nil
}

func saveETag(url string, resp *http.Response) error {
	err := os.MkdirAll(etagCacheDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create etag cache directory: %w", err)
	}

	etag := resp.Header.Get("etag")
	if etag == "" {
		return errNoETag
	}

	file, err := os.Create(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
	if err != nil {
		return fmt.Errorf("failed to create etag cache file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(etag)
	if err != nil {
		return fmt.Errorf("failed to write etag cache data: %w", err)
	}
	return nil
}

func loadETag(url string, req *http.Request) error {
	file, err := os.Open(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
	if err != nil {
		return errNoETag
	}
	defer file.Close()
	etag, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read etag: %w", err)
	}
	req.Header.Add("If-None-Match", string(etag))
	return nil
}
