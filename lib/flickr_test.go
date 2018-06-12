package flickr

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

var request = &Request{
	ApiKey: "",
	Args: map[string]string{
		"oauth_nonce":            "89601180",
		"oauth_timestamp":        "1528409127",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_token":            "",
	},
}

func TestSign(t *testing.T) {
	t.SkipNow()
	request.Method = "flickr.photosets.create"
	request.Args["title"] = "test_album"
	request.Args["primary_photo_id"] = ""
	request.sign(http.MethodPost, "https://api.flickr.com/services/rest", "")
	fmt.Println(request.Args["oauth_signature"])
	fmt.Println(url.QueryEscape(request.Args["oauth_signature"]))
}

func TestUpload(t *testing.T) {
	t.SkipNow()
	photoid, err := request.Upload("", "photo.jpg", "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(photoid)
}

func TestExecute(t *testing.T) {
	t.SkipNow()
	fmt.Println(request.Execute(http.MethodPost, ""))
}
