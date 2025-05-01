// pkg contains helper methods for collecting, parsing, validating the urls and more
package pkg

import (
	"bytes"
	"errors"
	"strings"

	"io"
	"net/http"
	"net/url"

	"github.com/vbauerster/mpb"
	_ "github.com/vbauerster/mpb/decor"
	"mvdan.cc/xurls/v2"
)

type FFDownloader struct {
	pageUrl     string
	downloadUrl string
	isValid     bool
	fileName    string
	progress    *mpb.Progress
}

func NewFF(pageUrl string, progress *mpb.Progress, spinner *mpb.Bar) (*FFDownloader, error) {
	ff := new(FFDownloader)
	ff.pageUrl = pageUrl
	ff.progress = progress
	return ff, ff.validateAndPrepareDownload(spinner)
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

func (ff *FFDownloader) Download() error {
	//bar := ff.progress.Add()
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
