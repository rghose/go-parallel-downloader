package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

var wg sync.WaitGroup

func getFileInfo(fileName string) (int,string) {
	res, err := http.Head(fileName)
	if err != nil {
		panic(err)
	}
	return int(res.ContentLength), res.Request.URL.String()
}

func download(url string, filePtr *os.File, from int, length int) error {
	fmt.Println("Getting range: ", from, " - ", from+length)
  chunkSize := (1024 * 1024)
	client := &http.Client{}
	for i:=0; i<length; i+=chunkSize {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
	  range_header := fmt.Sprintf("bytes=%d-%d", from+i, from+i+chunkSize)
		req.Header.Add("Range", range_header)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		filePtr.Write(responseData)
	}
	filePtr.Close()
	wg.Done()
	return nil
}

func main() {
	totalRoutines := 4
	tempFilePrefix := "dl_max"
	argsWithProg := os.Args

	if len(argsWithProg) < 2 {
		fmt.Println("Please specify a valid url to download")
		os.Exit(1)
	}

	toDownload := argsWithProg[1]

	fmt.Println("Downloading... ", toDownload)

	fileSize, realUrl := getFileInfo(toDownload)
	fmt.Println("Sizeof file = ", fileSize)

	wg.Add(totalRoutines)

	chunkSize := fileSize / totalRoutines

	for i := 0; i < totalRoutines; i++ {
		prefix := fmt.Sprintf("%d.%s", i+1, tempFilePrefix)
		file, err := ioutil.TempFile(os.TempDir(), prefix)
		if err != nil {
			fmt.Println("Could not create temporary files to download")
			os.Exit(2)
		}
		fmt.Println("Downloading to..", file.Name())
		go download(realUrl, file, (i * chunkSize), chunkSize)
	}

	wg.Wait()
	fmt.Println("Download complete")
}
