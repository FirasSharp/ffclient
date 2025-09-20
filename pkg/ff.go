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
	"path/filepath"
	"strings"

	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/vbauerster/mpb/v8"
	_ "github.com/vbauerster/mpb/v8/decor"
	"mvdan.cc/xurls/v2"
)

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
	if !ff.isValidSourceURL() {
		return errors.New("Invalid Url!")
	}

	body, err := ff.fetchPageContent()
	if err != nil {
		return err
	}

	downloadUrl, err := ff.extractDownloadURL(body)
	if err != nil {
		return err
	}
	ff.downloadUrl = downloadUrl
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

// isValidSourceURL checks if the page URL belongs to the expected domain.
// Returns true if the URL is valid, false otherwise.
func (ff *FFDownloader) isValidSourceURL() bool {
	link, err := url.ParseRequestURI(ff.pageUrl)
	if err != nil {
		return false
	}
	ffdomain, _ := url.ParseRequestURI("https://fuckingfast.co/")
	return (link.Host == ffdomain.Host) && (link.Scheme == ffdomain.Scheme)
}

// extractDownloadURL searches the page content for the download URL.
// It takes the page content as a byte slice and returns the found URL or an error.
func (ff *FFDownloader) extractDownloadURL(body []byte) (string, error) {
	rxRelaxed := xurls.Relaxed()
	foundUrls := rxRelaxed.FindAllString(string(body), -1)

	for _, foundUrl := range foundUrls {
		if strings.HasPrefix(foundUrl, "https://fuckingfast.co/dl/") {
			return foundUrl, nil
		}
	}

	return "", errors.New("Download url was not found!")
}

// fetchPageContent retrieves the content of the page URL.
// Returns the page content as a byte slice or an error if the request fails.
func (ff *FFDownloader) fetchPageContent() ([]byte, error) {
	resp, err := http.Get(ff.pageUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

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
