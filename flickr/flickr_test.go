package flickr

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

var auth = map[string]string{
	"oauth_token":        "",
	"oauth_consumer_key": ""}

var secret = ""

func TestSign(t *testing.T) {
	// t.SkipNow()
	additionalArgs := map[string]string{
		"method":           "flickr.photosets.create",
		"title":            "a b",
		"primary_photo_id": "40936585950",
	}
	request := NewRequest(http.MethodPost, auth, additionalArgs, secret)
	request.sign("https://api.flickr.com/services/rest")
	fmt.Println(request.args["oauth_signature"])
	fmt.Println(url.QueryEscape(request.args["oauth_signature"]))
}

func TestUpload(t *testing.T) {
	t.SkipNow()
	request := NewRequest(http.MethodPost, auth, nil, secret)
	photoid, err := request.Upload("photo.jpg")
	if err != nil {
		t.Fatalf("%+v", err)
	}
	fmt.Println(photoid)
}

func TestExecute(t *testing.T) {
	t.SkipNow()
	additionalArgs := map[string]string{
		"method":           "flickr.photosets.create",
		"title":            "test_title",
		"primary_photo_id": "40936585950",
	}
	request := NewRequest(http.MethodPost, auth, additionalArgs, secret)
	fmt.Println(request.Execute())
}
