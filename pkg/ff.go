// pkg contains helper methods for collecting, parsing, validating the urls and more
package pkg

import (
	"errors"
	"time"

	"github.com/vbauerster/mpb"
	_ "github.com/vbauerster/mpb/decor"
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
	return ff, ff.parseAndValidate(spinner)
}

func (ff *FF) parseAndValidate(spinner *mpb.Bar) error {
	defer spinner.Increment()
	if ff.url == "http://invalid" {
		return errors.New("Invalid Url!")
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (ff *FF) Download() error {
	//bar := ff.progress.Add()
	return nil
}
