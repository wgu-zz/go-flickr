package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/wgu/go-flickr/flickr"
)

func main() {
	requestTemplate, err := flickr.NewRequestFromCmd()
	flickr.CheckErr(err)

	args := map[string]string{
		"method": "flickr.photosets.getList",
	}
	request := flickr.NewRequest(http.MethodGet, requestTemplate.Auth, args, requestTemplate.Secret)
	response, err := request.ExecuteWithRetry(2, time.Second)
	flickr.CheckErr(err, response)
	var photoSets flickr.Photosets
	flickr.CheckErr(xml.Unmarshal([]byte(response), &photoSets), response)
	for _, photoSet := range photoSets.Photoset {
		folderName := photoSet.Title
		if existsFolder(folderName, "/Users/sgu/workspace/web-crawler/", "oumeirenti", "yazhourenti", "a4you", "hanguorenti", "ribenrenti", requestTemplate.Dir) {
			fmt.Println("Skipped " + folderName)
			continue
		}
		fmt.Println("Downloading " + folderName)
		folder := "/Users/sgu/workspace/web-crawler/" + requestTemplate.Dir + "/" + folderName + "/"
		os.MkdirAll(folder, os.ModePerm)
		for i := 1; ; i++ {
			args := map[string]string{
				"method":      "flickr.photosets.getPhotos",
				"user_id":     "161286677@N08",
				"photoset_id": photoSet.Id,
				"extras":      "url_o, original_format",
				"page":        strconv.Itoa(i),
			}
			request := flickr.NewRequest(http.MethodGet, requestTemplate.Auth, args, requestTemplate.Secret)
			response, err := request.ExecuteWithRetry(2, time.Second)
			flickr.CheckErr(err, response)
			var photoSet flickr.Photoset
			flickr.CheckErr(xml.Unmarshal([]byte(response), &photoSet), response)
			index := 1
			for _, photo := range photoSet.Photo {
				resp, err := http.Get(photo.UrlO)
				flickr.CheckErr(err)
				filePath := folder + photo.Title + "." + photo.OriginalFormat
				if exists(filePath) {
					filePath = folder + photo.Title + strconv.Itoa(index) + "." + photo.OriginalFormat
					index++
				}
				out, err := os.Create(filePath)
				flickr.CheckErr(err)
				_, err = io.Copy(out, resp.Body)
				flickr.CheckErr(err)
				resp.Body.Close()
				out.Close()
			}
			// time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			if i >= photoSet.Pages {
				break
			}
		}
	}
}

func existsFolder(folderName string, prefix string, paths ...string) bool {
	for _, path := range paths {
		if _, err := os.Stat(prefix + "/" + path + "/" + folderName); err == nil {
			return true
		} else if os.IsNotExist(err) {
			continue
		} else {
			panic(err)
		}
	}
	return false
}

func exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		panic(err)
	}
}
