// pkg contains helper methods for collecting, parsing, validating the urls and more
package pkg

import (
	"errors"
	"strings"

	"io"
	"net/http"
	"net/url"

	"github.com/vbauerster/mpb"
	_ "github.com/vbauerster/mpb/decor"
	"mvdan.cc/xurls/v2"
)

type FF struct {
	url         string
	downloadUrl string
	valid       bool
	fileName    string
	progress    *mpb.Progress
}

func NewFF(url string, progress *mpb.Progress, spinner *mpb.Bar) (*FF, error) {
	ff := new(FF)
	ff.url = url
	ff.progress = progress
	return ff, ff.checkURLAndExtractDownloadInfos(spinner)
}

func (ff *FF) checkURLAndExtractDownloadInfos(spinner *mpb.Bar) error {
	defer spinner.Increment()
	if !ff.validateUrl() {
		return errors.New("Invalid Url!")
	}

	body, err := ff.makeHTTPrequest()

	if err != nil {
		return err
	}

	downloadUrl, err := ff.extractDownloadLink(body)
	if err != nil {
		return err
	}
	ff.downloadUrl = downloadUrl
	ff.valid = true
	return nil
}

func (ff *FF) Download() error {
	//bar := ff.progress.Add()
	return nil
}

func (ff *FF) validateUrl() bool {
	link, err := url.ParseRequestURI(ff.url)
	if err != nil {
		return false
	}
	ffdomain, _ := url.ParseRequestURI("https://fuckingfast.co/")
	return (link.Host == ffdomain.Host) && (link.Scheme == ffdomain.Scheme)
}

func (ff *FF) extractDownloadLink(body []byte) (string, error) {
	rxRelaxed := xurls.Relaxed()
	urls := rxRelaxed.FindAllString(string(body), -1)

	for _, url := range urls {
		if strings.HasPrefix(url, "https://fuckingfast.co/dl/") {
			return url, nil
		}
	}

	return "", errors.New("Download url was not found!")
}

func (ff *FF) makeHTTPrequest() ([]byte, error) {
	resp, err := http.Get(ff.url)
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
