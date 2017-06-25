# go-parallel-downloader
Golang based parallel downloader

To run:

go run main.go -threads=6 -chunk=1024 url-to-download

Runs with 6 go threads (not real threads) and download chunk size of 1024 KB = 1 MB

