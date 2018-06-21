package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"

	flickr "github.com/wgu/go-flickr/lib"
)

var auth map[string]string
var secret string
var dir string

type photoset struct {
	Id string `xml:"id,attr"`
}

func main() {
	oauth_consumer_key := flag.String("oauth_consumer_key", "", "")
	oauth_token := flag.String("oauth_token", "", "")
	flag.StringVar(&secret, "secret", "", "")
	flag.StringVar(&dir, "dir", "", "")
	flag.Parse()
	auth = map[string]string{
		"oauth_consumer_key": *oauth_consumer_key,
		"oauth_token":        *oauth_token,
	}

	files, err := ioutil.ReadDir(dir)
	checkErr(err)
	var photosetid string
	var i int
	for _, fileinfo := range files {
		if mime.TypeByExtension(filepath.Ext(fileinfo.Name())) != "image/jpeg" {
			fmt.Println("Skipping non-jpeg file: " + fileinfo.Name())
			continue
		}
		fmt.Println("Uploading " + fileinfo.Name())
		photopath := filepath.Join(dir, fileinfo.Name())
		request := flickr.NewRequest(http.MethodPost, auth, nil, secret)
		photoid, err := request.UploadJpeg(photopath)
		checkErr(err)
		if i == 0 {
			additionalArgs := map[string]string{
				"method":           "flickr.photosets.create",
				"title":            filepath.Base(dir),
				"primary_photo_id": photoid,
			}
			request = flickr.NewRequest(http.MethodPost, auth, additionalArgs, secret)
			response, err := request.Execute()
			checkErr(err)
			fmt.Println(response)
			var pset photoset
			xml.Unmarshal([]byte(response), &pset)
			photosetid = pset.Id
			fmt.Println("Photaset id: " + photosetid)
		} else {
			additionalArgs := map[string]string{
				"method":      "flickr.photosets.addPhoto",
				"photoset_id": photosetid,
				"photo_id":    photoid,
			}
			request = flickr.NewRequest(http.MethodPost, auth, additionalArgs, secret)
			request.Execute()
		}
		i++
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
