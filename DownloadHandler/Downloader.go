package DownloadHandler

import (
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func GetDownloadResponse(downloadRequest DownloadRequest) string {

	if isNonValidType(downloadRequest.Type) {
		return ""
	}

	currID := uuid.New().String()
	currStatus := new(DownloadStatus)
	DownloadCollection[currID] = currStatus

	createDirectory(GLOBAL_PATH + currID)

	if downloadRequest.Type == "serial" {
		serialDownloader(downloadRequest.Urls, currID)
	} else {
		go concurrentDownloader(downloadRequest.Urls, currID)
	}

	return currID
}

func serialDownloader(urls []string, currID string) {

	DownloadCollection[currID].initializeStatus(currID, "SERIAL", "QUEUED")

	for _, v := range urls {

		destination := GLOBAL_PATH + currID + "/" + getFileName(v)
		err := downloadFile(v, destination)
		if err != nil {
			DownloadCollection[currID].Status = "FAILED"
			DownloadCollection[currID].EndTime = time.Now()
			log.Fatal(err)
		}
		DownloadCollection[currID].Files[v] = destination
	}

	DownloadCollection[currID].EndTime = time.Now()
	DownloadCollection[currID].Status = "SUCCESSFUL"
}

func concurrentDownloader(urls []string, currID string) {

	DownloadCollection[currID].initializeStatus(currID, "CONCURRENT", "QUEUED")
	var wg sync.WaitGroup

	for _, v := range urls {

		destination := GLOBAL_PATH + currID + "/" + getFileName(v)
		wg.Add(1)
		go concurrentDownloadHelper(v, destination, currID, &wg)
		DownloadCollection[currID].Files[v] = destination
	}
	wg.Wait()

	DownloadCollection[currID].EndTime = time.Now()
	DownloadCollection[currID].Status = "SUCCESSFUL"
}

func concurrentDownloadHelper(url, filepath, currID string, wg *sync.WaitGroup) {
	err := downloadFile(url, filepath)
	if err != nil {
		DownloadCollection[currID].Status = "FAILED"
		DownloadCollection[currID].EndTime = time.Now()
		log.Fatal(err)
	}
	wg.Done()
}

func downloadFile(url string, filepath string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}