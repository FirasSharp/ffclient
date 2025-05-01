//MIT License

//Copyright (c) 2025 Firas Jelassi

//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

// pkg contains helper methods for collecting, parsing, validating the urls and more
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

type FFDownloader struct {
	pageUrl     string
	downloadUrl string
	isValid     bool
	fileName    string
}

func NewFF(pageUrl string, spinner *mpb.Bar) (*FFDownloader, error) {
	ff := new(FFDownloader)
	ff.pageUrl = pageUrl
	return ff, ff.validateAndPrepareDownload(spinner)
}

// getters

func (ff *FFDownloader) PageUrl() string {
	return ff.pageUrl
}

func (ff *FFDownloader) IsValid() bool {
	return ff.isValid
}

func (ff *FFDownloader) DownloadUrl() string {
	return ff.downloadUrl
}

func (ff *FFDownloader) FileName() string {
	return ff.fileName
}

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
	io.Copy(dest, reader)
	return nil
}

func (ff *FFDownloader) isValidSourceURL() bool {
	link, err := url.ParseRequestURI(ff.pageUrl)
	if err != nil {
		return false
	}
	ffdomain, _ := url.ParseRequestURI("https://fuckingfast.co/")
	return (link.Host == ffdomain.Host) && (link.Scheme == ffdomain.Scheme)
}

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

func (ff *FFDownloader) getFileName(body []byte) (string, error) {
	title, _, err := GetHtmlTitle(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return title, nil
}
