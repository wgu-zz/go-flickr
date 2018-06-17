package flickr

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

var request *Request = NewRequest(map[string]string{
	"oauth_token":        "",
	"oauth_consumer_key": ""})

func TestSign(t *testing.T) {
	t.SkipNow()
	request.Args["method"] = "flickr.photosets.create"
	request.Args["title"] = "test_album"
	request.Args["primary_photo_id"] = ""
	request.sign("https://api.flickr.com/services/rest", "")
	fmt.Println(request.Args["oauth_signature"])
	fmt.Println(url.QueryEscape(request.Args["oauth_signature"]))
}

func TestUpload(t *testing.T) {
	t.SkipNow()
	request.HttpMethod = http.MethodPost
	photoid, err := request.Upload("", "image/jpeg", "")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	fmt.Println(photoid)
}

func TestExecute(t *testing.T) {
	t.SkipNow()
	fmt.Println(request.Execute(""))
}
