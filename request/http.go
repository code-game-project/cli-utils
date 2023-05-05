package request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"

	"github.com/code-game-project/cli-utils/cli"
	"github.com/code-game-project/cli-utils/feedback"

	neturl "net/url"
)

const FeedbackPkg = feedback.Package("request")

var (
	httpCacheDir = filepath.Join(xdg.CacheHome, "codegame", "http")
	etagCacheDir = filepath.Join(httpCacheDir, "etag")
)

var errNoETag = errors.New("no etag")

type reader struct {
	r           io.ReadCloser
	w           io.WriteCloser
	bytesRead   int64
	contentSize int64
	url         string
	method      string
	onErr       func(err error)
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if n > 0 {
		if r.contentSize > 0 {
			r.bytesRead += int64(n)
			feedback.Progress(FeedbackPkg, fmt.Sprintf("fetch %s %s", r.method, r.url), fmt.Sprintf("Fetching %s %s", r.method, r.url), int64(r.bytesRead), r.contentSize, cli.UnitFileSize)
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
	if errors.Is(err, io.EOF) {
		if r.contentSize > 0 {
			feedback.Progress(FeedbackPkg, fmt.Sprintf("fetch %s %s", r.method, r.url), fmt.Sprintf("Fetching %s %s", r.method, r.url), r.contentSize, r.contentSize, cli.UnitFileSize)
		}
		return n, io.EOF
	}
	if err != nil && !errors.Is(err, io.EOF) {
		if r.w != nil {
			r.w.Close()
			r.w = nil
		}
		r.onErr(err)
	}
	return n, err
}

func (r *reader) Close() error {
	if r.w != nil {
		r.w.Close()
	}
	return r.r.Close()
}

func Fetch(url, method string, cacheMaxAge time.Duration, timeout time.Duration, reportProgress bool, body io.Reader) (responseBody io.ReadCloser, statusCode int, err error) {
	feedback.Debug(FeedbackPkg, "Fetching %s %s...", strings.ToUpper(method), url)
	cacheFilePath := filepath.Join(httpCacheDir, neturl.PathEscape(url))
	if cacheMaxAge > 0 {
		if stat, err := os.Stat(cacheFilePath); err == nil && time.Since(stat.ModTime()) <= cacheMaxAge {
			file, err := os.Open(cacheFilePath)
			if err == nil {
				feedback.Debug(FeedbackPkg, "Found in cache.")
				return file, 0, nil
			}
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("create http request: %w", err)
	}
	if _, err = os.Stat(cacheFilePath); err == nil {
		loadETag(url, req)
	}
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == http.StatusNotModified {
		file, err2 := os.Open(cacheFilePath)
		if err2 == nil {
			if err == nil {
				os.Chtimes(cacheFilePath, time.Now(), time.Now())
				feedback.Debug(FeedbackPkg, "Sever returned 304 Not Modified. Using cached version.")
			} else {
				feedback.Debug(FeedbackPkg, "Offline. Using cached version.")
			}
			return file, 0, nil
		}
		os.Remove(filepath.Join(etagCacheDir, neturl.PathEscape(url)))
		return nil, 0, fmt.Errorf("fetch data: %w", err)
	}
	if resp.StatusCode >= 300 {
		statusCode = resp.StatusCode
		cacheMaxAge = 0
	} else {
		saveETag(url, resp)
	}

	var cache io.WriteCloser
	if cacheMaxAge > 0 {
		err := os.MkdirAll(httpCacheDir, 0o755)
		if err != nil {
			return nil, 0, fmt.Errorf("create cache directory: %w", err)
		}
		file, err := os.Create(cacheFilePath)
		if err != nil {
			return nil, 0, fmt.Errorf("create cache file: %w", err)
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
		onErr: func(_ error) {
			if cache != nil {
				os.Remove(cacheFilePath)
			}
		},
	}, statusCode, nil
}

func FetchFile(url string, cacheMaxAge time.Duration, reportProgress bool) (io.ReadCloser, error) {
	body, status, err := Fetch(url, "GET", cacheMaxAge, 0, reportProgress, nil)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		body.Close()
		return nil, fmt.Errorf("http status: %s", http.StatusText(status))
	}
	return body, nil
}

func FetchJSON[T any](url string, maxCacheAge time.Duration) (T, error) {
	var obj T
	file, status, err := Fetch(url, "GET", maxCacheAge, 10*time.Second, false, nil)
	if err != nil && !errors.Is(err, io.EOF) {
		return obj, err
	}
	defer file.Close()
	if status >= 300 {
		return obj, fmt.Errorf("http status: %s", http.StatusText(status))
	}
	err = json.NewDecoder(file).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("decode response from '%s': %w", url, err)
	}
	return obj, nil
}

func PostJSON[T any](url string, data any) (T, error) {
	var obj T
	buf, err := json.Marshal(data)
	if err != nil {
		return obj, fmt.Errorf("marshal json payload for '%s': %w", url, err)
	}
	file, status, err := Fetch(url, "POST", 0, 10*time.Second, false, bytes.NewBuffer(buf))
	if err != nil {
		return obj, err
	}
	defer file.Close()
	if status >= 300 {
		return obj, fmt.Errorf("http status: %s", http.StatusText(status))
	}
	err = json.NewDecoder(file).Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("decode response from '%s': %w", url, err)
	}
	return obj, nil
}

// TrimURL removes the protocol component and trailing slashes.
func TrimURL(url string) string {
	u, err := neturl.Parse(url)
	if err != nil {
		return url
	}
	u.Scheme = ""
	return strings.TrimSuffix(u.String(), "/")
}

// BaseURL prepends `protocol + "://"` or `protocol + "s://"` to the url depending on TLS support.
func BaseURL(protocol string, trimmedURL string, a ...any) string {
	trimmedURL = fmt.Sprintf(trimmedURL, a...)
	if IsTLS(trimmedURL) {
		return protocol + "s://" + trimmedURL
	} else {
		return protocol + "://" + trimmedURL
	}
}

var isTLSCache = make(map[string]bool, 0)

// IsTLS verifies the TLS certificate of a trimmed URL.
func IsTLS(trimmedURL string) (isTLS bool) {
	if is, ok := isTLSCache[trimmedURL]; ok {
		return is
	}
	defer func() {
		isTLSCache[trimmedURL] = isTLS
	}()
	url, err := neturl.Parse("https://" + trimmedURL)
	if err != nil {
		return false
	}
	host := url.Host
	if url.Port() == "" {
		host = host + ":443"
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", host, &tls.Config{})
	if err != nil {
		return false
	}
	defer conn.Close()

	err = conn.VerifyHostname(url.Hostname())
	if err != nil {
		return false
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	return !time.Now().After(expiry)
}

// HasContentType returns true if the 'content-type' header includes mimetype.
func HasContentType(h http.Header, mimetype string) bool {
	contentType := h.Get("content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
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
