package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var wg sync.WaitGroup

func getFileInfo(fileName string) (int, string) {
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
	for i := 0; i < length; i += chunkSize {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		upto := from + i + chunkSize
		if i+chunkSize > length {
			upto = (from + length)
			fmt.Printf("%d, %d, %d. upto = %d\n", i, chunkSize, length, upto)
		}
		range_header := fmt.Sprintf("bytes=%d-%d", from+i, upto-1)
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
	totalRoutines := 6
	tempFilePrefix := "dlmax."
	argsWithProg := os.Args

	if len(argsWithProg) < 2 {
		fmt.Println("Please specify a valid url to download")
		os.Exit(1)
	}

	toDownload := argsWithProg[1]
	destFileName := filepath.Base(toDownload)

	fmt.Println("Downloading... ", toDownload)

	fileSize, realUrl := getFileInfo(toDownload)
	fmt.Println("Sizeof file = ", fileSize)

	wg.Add(totalRoutines)

	chunkSize := fileSize / totalRoutines
	extraChunk := fileSize % totalRoutines

	files := make([]string, totalRoutines)

	for i := 0; i < totalRoutines; i++ {
		prefix := fmt.Sprintf("%d.%s", i+1, tempFilePrefix)
		file, err := ioutil.TempFile(os.TempDir(), prefix)
		if err != nil {
			fmt.Println("Could not create temporary files to download")
			os.Exit(2)
		}
		fmt.Println("Downloading to..", file.Name())
		if totalRoutines-1 == i {
			go download(realUrl, file, (i * chunkSize), chunkSize+extraChunk)
		} else {
			go download(realUrl, file, (i * chunkSize), chunkSize)
		}
		files[i] = file.Name()
	}

	wg.Wait()
	fmt.Println("Downloaded all parts")

	out, err := os.Create(destFileName)
	if err != nil {
		fmt.Println("Could not open destination file")
		os.Exit(3)
	}
	defer out.Close()
	for i := 0; i < totalRoutines; i++ {
		in, err := os.Open(files[i])
		if err != nil {
			fmt.Println("Bad shit.")
		}
		io.Copy(out, in)
		in.Close()
		os.Remove(files[i])
	}
	fmt.Println("Download success!")
}
