// MIT License
//
// Copyright (c) 2025 Firas Jelassi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package pkg contains the ffclient and helper methods.
package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/vbauerster/mpb/v8"
	_ "github.com/vbauerster/mpb/v8/decor"
)

var jar, _ = cookiejar.New(nil)
var client *http.Client = &http.Client{
	Jar: jar,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: 30 * time.Second,
}

// FFDownloader represents a downloader for resources from https://fuckingfast.co/
// It handles URL validation, content fetching, and file downloading.
type FFDownloader struct {
	pageUrl     string
	downloadUrl string
	isValid     bool
	fileName    string
}

// NewFF creates a new FFDownloader instance and validates the provided page URL.
// It takes a page URL string and a progress bar spinner as parameters.
// Returns a new FFDownloader instance and an error if validation fails.
func NewFF(pageUrl string, spinner *mpb.Bar) (*FFDownloader, error) {
	ff := new(FFDownloader)
	ff.pageUrl = pageUrl
	return ff, ff.validateAndPrepareDownload(spinner)
}

// PageUrl returns the page URL associated with the FFDownloader.
func (ff *FFDownloader) PageUrl() string {
	return ff.pageUrl
}

// IsValid returns whether the FFDownloader has a valid configuration.
func (ff *FFDownloader) IsValid() bool {
	return ff.isValid
}

// DownloadUrl returns the download URL extracted from the page.
func (ff *FFDownloader) DownloadUrl() string {
	return ff.downloadUrl
}

// FileName returns the filename extracted from the page content.
func (ff *FFDownloader) FileName() string {
	return ff.fileName
}

// validateAndPrepareDownload validates the source URL, fetches page content,
// extracts download URL and filename. It takes a progress bar spinner as parameter.
// Returns an error if any step in the validation process fails.
func (ff *FFDownloader) validateAndPrepareDownload(spinner *mpb.Bar) error {
	defer spinner.Increment()

	sourcePath, ok := ff.validateAndGetSourcePath()
	if !ok {
		return errors.New("Invalid Url!")
	}

	err := ff.extractDownloadURL(sourcePath)
	if err != nil {
		return err
	}

	body, err := ff.fetchPageContent()

	ff.isValid = true
	fileName, err := ff.getFileName(body)
	if err != nil {
		return err
	}

	ff.fileName = fileName

	return nil
}

// Download performs the actual file download to the specified path.
// It takes a destination path and a progress bar as parameters.
// Returns an error if the download fails at any stage.
func (ff *FFDownloader) Download(path string, bar *mpb.Bar) error {
	url := ff.DownloadUrl()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	destName := filepath.Join(path, ff.FileName())
	dest, err := os.Create(destName)
	if err != nil {
		return err
	}
	defer dest.Close()

	reader := bar.ProxyReader(resp.Body)

	_, err = io.Copy(dest, reader)
	if err != nil {
		return err
	}

	return nil
}

// validateAndGetSourcePath checks if the page URL belongs to the expected domain.
// Returns true if the URL is valid, false otherwise and the url path
func (ff *FFDownloader) validateAndGetSourcePath() (string, bool) {
	link, err := url.ParseRequestURI(ff.pageUrl)
	if err != nil {
		return "", false
	}
	ffdomain, _ := url.ParseRequestURI("https://fuckingfast.co/")
	sourcePath := strings.Split(link.Path, "#")[0]
	return sourcePath, (link.Host == ffdomain.Host) && (link.Scheme == ffdomain.Scheme)
}

// extractDownloadURL searches the page content for the download URL.
// It takes the page content as a byte slice and returns the found URL or an error.
func (ff *FFDownloader) extractDownloadURL(sourcePath string) error {
	pageUrl := "https://fuckingfast.co/f%s/go"
	pageUrl = fmt.Sprintf(pageUrl, sourcePath)
	req, err := http.NewRequest("POST", pageUrl, nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.7")
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/157.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ff.downloadUrl = resp.Header.Get("Hx-Redirect")

	return nil
}

// fetchPageContent retrieves the content of the page URL.
// Returns the page content as a byte slice or an error if the request fails.
func (ff *FFDownloader) fetchPageContent() ([]byte, error) {

	req, err := http.NewRequest("GET", ff.pageUrl, nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.7")
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/157.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)

	return bytes, nil
}

// getFileName extracts the filename from the page content.
// It takes the page content as a byte slice and returns the filename or an error.
func (ff *FFDownloader) getFileName(body []byte) (string, error) {
	title, _, err := GetHtmlTitle(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return title, nil
}
