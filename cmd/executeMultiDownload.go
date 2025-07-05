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

// Package cmd implements the command-line interface for the ffclient application,
// handling the multi-download functionality from https://fuckingfast.co/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/FirasSharp/ffclient/pkg"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"net/http"
)

// failureInfo contains information about a failed download attempt
type failureInfo struct {
	url string // The URL that failed to download
	err error  // The error that occurred
}

// failed tracks all download failures in a thread-safe manner
type failed struct {
	infos []failureInfo // Slice of failure information
	sy    sync.Mutex    // Mutex for thread-safe operations
}

var failure = &failed{}

// ExecuteMultiDownload coordinates the multi-file download process.
// It takes a save path, input file path, and comma-separated links as input.
// Returns an error if the download process fails completely.
func ExecuteMultiDownload(savePath, inputFile, links string) error {
	wg := new(sync.WaitGroup)
	spinnerProgress := mpb.New(mpb.WithWaitGroup(wg))

	urls, err := getUrls(inputFile, links)
	if err != nil {
		return err
	}

	ffs := createFF(urls, wg, spinnerProgress)
	ffs = filterSlice(ffs)
	download(ffs, wg, savePath)

	if len(failure.infos) == 0 {
		log.Entry().Info("All files were downloaded successfully!")
		return nil
	}

	if len(failure.infos) == len(urls) {
		log.Entry().Error("No file was downloaded!")
	} else {
		log.Entry().Errorf("%d out of %d were successfully downloaded!", len(urls)-len(failure.infos), len(urls))
	}

	for _, info := range failure.infos {
		log.Entry().Errorf("Failed to process link '%s': %v", info.url, info.err)
	}

	return nil
}

// createFF initializes FFDownloader instances for each URL with progress tracking.
// Takes a slice of URLs, a wait group, and a progress bar container.
// Returns a slice of initialized FFDownloader instances.
func createFF(urls []string, wg *sync.WaitGroup, spinnerProgress *mpb.Progress) []*pkg.FFDownloader {
	ffs := make([]*pkg.FFDownloader, 0)

	spinner := spinnerProgress.AddSpinner(int64(len(urls)), mpb.PrependDecorators(
		decor.Name("Validating URL & finding download link", decor.WC{W: len("Validating URL & finding download link") + 1, C: decor.DindentRight}),
		decor.Counters(0, " %d / %d")),
		mpb.AppendDecorators(decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), "done")))

	wg.Add(len(urls))
	for _, url := range urls {
		go func() {
			defer wg.Done()
			ff, err := pkg.NewFF(url, spinner)
			if err != nil {
				failure.sy.Lock()
				defer failure.sy.Unlock()
				info := failureInfo{
					url: url,
					err: err,
				}
				failure.infos = append(failure.infos, info)
			}
			ffs = append(ffs, ff)
		}()
	}

	spinnerProgress.Wait()
	return ffs
}

// download performs the actual file downloads with progress tracking.
// Takes a slice of FFDownloader instances, a wait group, and the save path.
func download(ffs []*pkg.FFDownloader, wg *sync.WaitGroup, savePath string) {
	downloadProgressBar := mpb.New(mpb.WithWaitGroup(wg))
	wg.Add(len(ffs))
	for _, ff := range ffs {
		size, err := getFileSize(ff.DownloadUrl())
		if err != nil {
			failure.sy.Lock()
			info := failureInfo{
				url: ff.PageUrl(),
				err: err,
			}
			failure.infos = append(failure.infos, info)
			failure.sy.Unlock()
			continue
		}
		bar := downloadProgressBar.AddBar(size,
			mpb.PrependDecorators(
				decor.CountersKibiByte("% 6.1f / % 6.1f"),
			),
			mpb.AppendDecorators(
				decor.EwmaETA(decor.ET_STYLE_MMSS, float64(size)/2048),
				decor.Name(" ] "),
				decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 30),
				decor.OnComplete(
					decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
				),
			),
		)

		go func() {
			defer wg.Done()
			err := ff.Download(savePath, bar)
			if err != nil {
				failure.sy.Lock()
				defer failure.sy.Unlock()
				info := failureInfo{
					url: ff.PageUrl(),
					err: err,
				}
				failure.infos = append(failure.infos, info)
			}
		}()
	}
	downloadProgressBar.Wait()
	log.Entry().Info("Download Completed!")
}

// filterSlice removes any invalid FFDownloader instances from the slice.
// Takes a slice of FFDownloader instances and returns a filtered slice.
func filterSlice(ff []*pkg.FFDownloader) []*pkg.FFDownloader {
	res := make([]*pkg.FFDownloader, 0)
	for _, f := range ff {
		if f.IsValid() {
			res = append(res, f)
		}
	}
	return res
}

// getUrls retrieves URLs either from a file or from a comma-separated string.
// Takes an input file path and a comma-separated string of links.
// Returns a slice of URLs or an error if reading fails.
func getUrls(inputFile, links string) ([]string, error) {
	if len(inputFile) > 0 {
		return getUrlsFromFile(inputFile)
	}
	return getUrlsFromString(links)
}

// getUrlsFromFile reads URLs from a text file (one URL per line).
// Takes a file path and returns a slice of URLs or an error if reading fails.
func getUrlsFromFile(inputFilePath string) ([]string, error) {
	result := make([]string, 0)
	file, err := os.Open(inputFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	s := bufio.NewScanner(file)
	for s.Scan() {
		result = append(result, s.Text())
	}
	return result, nil
}

// getUrlsFromString splits a comma-separated string into individual URLs.
// Takes a comma-separated string and returns a slice of URLs.
func getUrlsFromString(links string) ([]string, error) {
	urls := strings.Split(links, ",")
	return urls, nil
}

// getFileSize retrieves the size of a remote file via HTTP HEAD request.
// Takes a URL and returns the file size in bytes or an error if the request fails.
func getFileSize(url string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("server returned: %s", resp.Status)
	}

	if resp.ContentLength > 0 {
		return resp.ContentLength, nil
	}

	return -1, fmt.Errorf("Content-Length not provided")
}
