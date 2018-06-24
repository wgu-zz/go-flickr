package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/wgu/go-flickr/flickr"
)

type photoset struct {
	Id string `xml:"id,attr"`
}

func main() {
	requestTemplate, err := flickr.NewRequestFromCmd()
	checkErr(err)

	files, err := ioutil.ReadDir(requestTemplate.Dir)
	checkErr(err)
	var photosetid string
	var i int
	for _, fileinfo := range files {
		if mime.TypeByExtension(filepath.Ext(fileinfo.Name())) != "image/jpeg" {
			fmt.Println("Skipping non-jpeg file: " + fileinfo.Name())
			continue
		}
		fmt.Println("Uploading " + fileinfo.Name())
		photopath := filepath.Join(requestTemplate.Dir, fileinfo.Name())
		request := flickr.NewRequest(http.MethodPost, requestTemplate.Auth, nil, requestTemplate.Secret)
		photoid, err := request.UploadJpeg(photopath)
		checkErr(err)
		if i == 0 {
			fmt.Println("Creating album")
			additionalArgs := map[string]string{
				"method":           "flickr.photosets.create",
				"title":            filepath.Base(requestTemplate.Dir),
				"primary_photo_id": photoid,
			}
			request = flickr.NewRequest(http.MethodPost, requestTemplate.Auth, additionalArgs, requestTemplate.Secret)
			response, err := request.Execute()
			checkErr(err)
			// fmt.Println(response)
			var pset photoset
			xml.Unmarshal([]byte(response), &pset)
			photosetid = pset.Id
			fmt.Println("Photaset id: " + photosetid)
		} else {
			fmt.Println("Adding " + photoid + " to album")
			additionalArgs := map[string]string{
				"method":      "flickr.photosets.addPhoto",
				"photoset_id": photosetid,
				"photo_id":    photoid,
			}
			request = flickr.NewRequest(http.MethodPost, requestTemplate.Auth, additionalArgs, requestTemplate.Secret)
			request.Execute()
		}
		i++
	}
	if requestTemplate.CollectionId != "" {
		fmt.Println("Adding album " + photosetid + " to collection")
		additionalArgs := map[string]string{
			"method":        "flickr.collections.addSet",
			"collection_id": requestTemplate.CollectionId,
			"photoset_id":   photosetid,
		}
		request := flickr.NewRequest(http.MethodPost, requestTemplate.Auth, additionalArgs, requestTemplate.Secret)
		request.Execute()
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
