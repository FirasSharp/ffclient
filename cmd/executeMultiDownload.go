package cmd

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/FirasSharp/ffclient/pkg"
	"github.com/SAP/jenkins-library/pkg/log"
	_ "github.com/SAP/jenkins-library/pkg/log"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type failureInfo struct {
	url string
	err error
}

type failed struct {
	infos []failureInfo
	sy    sync.Mutex
}

func ExecuteMultiDownload(savePath, inputFile, links string) error {
	ffs := make([]*pkg.FF, 0)
	failed := failed{}
	wg := new(sync.WaitGroup)
	spinnerProgress := mpb.New(mpb.WithWaitGroup(wg))
	downloadProgressBar := mpb.New(mpb.WithWaitGroup(wg))

	urls, err := getUrls(inputFile, links)

	if err != nil {
		return err
	}

	spinner := spinnerProgress.AddSpinner(int64(len(urls)), mpb.SpinnerOnLeft, mpb.PrependDecorators(
		decor.Name("Validating URL & finding download link", decor.WC{W: len("Validating URL & finding download link") + 1, C: decor.DidentRight}),
		decor.Counters(0, " %d / %d")),
		mpb.AppendDecorators(decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_GO, 60), "done")))

	wg.Add(len(urls))
	for _, url := range urls {
		go func() {
			defer wg.Done()
			ff, err := pkg.NewFF(url, downloadProgressBar, spinner)
			if err != nil {
				failed.sy.Lock()
				defer failed.sy.Unlock()
				info := failureInfo{
					url: url,
					err: err,
				}
				failed.infos = append(failed.infos, info)
			}
			ffs = append(ffs, ff)
		}()
	}

	spinnerProgress.Wait()

	if len(failed.infos) == 0 {
		log.Entry().Info("All files were downloaded successfully!")
		return nil
	}

	if len(failed.infos) == len(urls) {
		log.Entry().Error("No file was downloaded!")
	} else {
		log.Entry().Errorf("%d out of %d were successfully downloaded!", len(urls)-len(failed.infos), len(urls))
	}

	for _, info := range failed.infos {
		log.Entry().Errorf("Failed to process link '%s': %v", info.url, info.err)
	}

	return nil
}

func getUrls(inputFile, links string) ([]string, error) {
	if len(inputFile) > 0 {
		return getUrlsFromFile(inputFile)
	}
	return getUrlsFromString(links)
}

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

func getUrlsFromString(links string) ([]string, error) {
	urls := strings.Split(links, ",")
	return urls, nil
}
