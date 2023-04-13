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

	"github.com/code-game-project/cli-utils/feedback"

	neturl "net/url"
)

const FeedbackPkg = feedback.Package("request")

var (
	httpCacheDir = filepath.Join(xdg.CacheHome, "codegame", "http")
	etagCacheDir = filepath.Join(httpCacheDir, "etag")
)

var errNoETag = errors.New("no etag")

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type reader struct {
	r            io.ReadCloser
	w            io.WriteCloser
	bytesRead    int
	lastByteRead int
	contentSize  int64
	url          string
	onErr        func(err error)
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		if r.contentSize > 0 {
			r.bytesRead += n
			if r.lastByteRead == 0 || int64(r.bytesRead) == r.contentSize || r.bytesRead-r.lastByteRead > 100000 {
				r.lastByteRead = r.bytesRead
				feedback.Progress(FeedbackPkg, fmt.Sprintf("fetch %s", r.url), fmt.Sprintf("Fetching %s (%.2fkB/%.2fkB)...", r.url, float64(r.bytesRead)/1000, float64(r.contentSize)/1000), float64(r.bytesRead), float64(r.contentSize))
			}
		}
		if r.w != nil {
			if n2, err2 := r.w.Write(p[:n]); err2 != nil {
				r.w.Close()
				r.w = nil
				r.onErr(err2)
				return n2, err2
			}
		}
	}
	if err != nil && !errors.Is(err, io.EOF) {
		if r.w != nil {
			r.w.Close()
			r.w = nil
		}
		r.onErr(err)
	}
	return
}

func (r *reader) Close() error {
	if r.w != nil {
		r.w.Close()
	}
	return r.r.Close()
}

func FetchFile(url string, cacheMaxAge time.Duration, reportProgress bool) (io.ReadCloser, error) {
	feedback.Debug(FeedbackPkg, "Fetching %s...", url)
	cacheFilePath := filepath.Join(httpCacheDir, neturl.PathEscape(url))
	if cacheMaxAge > 0 {
		if stat, err := os.Stat(cacheFilePath); err == nil && time.Since(stat.ModTime()) <= cacheMaxAge {
			file, err := os.Open(cacheFilePath)
			if err == nil {
				feedback.Debug(FeedbackPkg, "Found in cache.")
				return file, nil
			}
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	if _, err = os.Stat(cacheFilePath); err == nil {
		loadETag(url, req)
	}
	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode == http.StatusNotModified {
		file, err2 := os.Open(cacheFilePath)
		if err2 == nil {
			if err == nil {
				os.Chtimes(cacheFilePath, time.Now(), time.Now())
				feedback.Debug(FeedbackPkg, "Sever returned 304 Not Modified. Using cached version.")
			} else {
				feedback.Debug(FeedbackPkg, "Offline. Using cached version.")
			}
			return file, nil
		}
		os.Remove(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
		return nil, fmt.Errorf("fetch data: %w", err)
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch '%s': status '%s'", url, resp.Status)
	}
	saveETag(url, resp)

	var cache io.WriteCloser
	if cacheMaxAge > 0 {
		err := os.MkdirAll(httpCacheDir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("create cache directory: %w", err)
		}
		file, err := os.Create(cacheFilePath)
		if err != nil {
			return nil, fmt.Errorf("create cache file: %w", err)
		}
		cache = file
	}

	contentSize := resp.ContentLength
	if !reportProgress {
		contentSize = 0
	}

	return &reader{
		r:           resp.Body,
		w:           cache,
		url:         url,
		contentSize: contentSize,
		onErr: func(err error) {
			fmt.Println(err)
			if cache != nil {
				os.Remove(cacheFilePath)
			}
		},
	}, nil
}

func FetchJSON[T any](url string, maxCacheAge time.Duration) (T, error) {
	var obj T
	file, err := FetchFile(url, maxCacheAge, false)
	if err != nil {
		return obj, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("decode response from '%s': %w", url, err)
	}
	return obj, nil
}

func saveETag(url string, resp *http.Response) error {
	err := os.MkdirAll(etagCacheDir, 0o755)
	if err != nil {
		return fmt.Errorf("create etag cache directory: %w", err)
	}

	etag := resp.Header.Get("etag")
	if etag == "" {
		return errNoETag
	}

	file, err := os.Create(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
	if err != nil {
		return fmt.Errorf("create etag cache file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(etag)
	if err != nil {
		return fmt.Errorf("write etag cache data: %w", err)
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
		return fmt.Errorf("read etag: %w", err)
	}
	req.Header.Add("If-None-Match", string(etag))
	return nil
}
