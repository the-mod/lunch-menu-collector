package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	urlFileName = "urls.txt"
)

type Download struct {
	FileName string
	Content  []byte
}

func getNumberOfWeek() int {
	now := time.Now().UTC()
	_, week := now.ISOWeek() // _ stands for year
	return week
}

func getFileName(urlString string) string {
	url, err := url.Parse(urlString)
	if err != nil {
		log.Fatal(err)
	}
	parts := strings.Split(urlString, "/")
	// TODO: what if last part is empty? like with html page?
	docName := parts[len(parts)-1]
	fileName := fmt.Sprintf("%s__KW%d__%s", url.Host, getNumberOfWeek(), docName)
	return fileName
}

func makeDirIfNotExists() string {
	dirName := fmt.Sprintf("KW%d", getNumberOfWeek())
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		os.Mkdir(dirName, 666)
	}
	return dirName
}

func createRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	return req
}

func doRequest(url string) ([]byte, error) {
	req := createRequest(url)
	client := &http.Client{}
	resp, _ := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return body, err
}

func saveFile(fileName string, payload []byte) {
	ioutil.WriteFile(fileName, payload, 666)
}

func readFile() []string {
	var urls []string
	file, err := os.Open(urlFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return urls
}

func download(url string, fileName string, ch chan *Download) {
	body, err := doRequest(url)

	if err != nil {
		log.Fatal(err)
	}
	download := &Download{
		FileName: fileName,
		Content:  body,
	}
	ch <- download
}

func main() {
	urls := readFile()
	fmt.Println(urls)
	dirName := makeDirIfNotExists()
	ch := make(chan *Download)
	var wg sync.WaitGroup
	wg.Add(len(urls))
	defer wg.Wait()

	for _, url := range urls {
		fileName := getFileName(url)
		go download(url, fileName, ch)
	}

	go func() {
		for downloads := range ch {
			saveFile(dirName+"/"+downloads.FileName, downloads.Content)
			fmt.Println("Saved " + downloads.FileName)
			wg.Done()
		}
	}()
}
