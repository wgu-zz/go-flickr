package flickr

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	endpoint        = "https://api.flickr.com/services/rest"
	uploadEndpoint  = "https://up.flickr.com/services/upload"
	replaceEndpoint = "https://up.flickr.com/services/replace"
	image_jpeg      = "image/jpeg"
)

type Request struct {
	ApiKey string
	Method string
	Args   map[string]string
}

type Response struct {
	Status  string         `xml:"stat,attr"`
	Error   *ResponseError `xml:"err"`
	Payload string         `xml:",innerxml"`
}

type ResponseError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:"msg,attr"`
}

// type nopCloser struct {
// 	io.Reader
// }

// func (nopCloser) Close() error { return nil }

type Error string

func (e Error) Error() string {
	return string(e)
}

func (request *Request) sign(verb string, request_url string, secret string) {
	args := request.Args

	delete(args, "oauth_signature")

	sorted_keys := make([]string, len(args)+2)

	args["oauth_consumer_key"] = request.ApiKey
	args["method"] = request.Method

	// Sort array keys
	i := 0
	for k := range args {
		sorted_keys[i] = k
		i++
	}
	sort.Strings(sorted_keys)

	// Build out ordered key-value string prefixed by secret
	base := verb + "&" + url.QueryEscape(request_url) + "&"
	var params string
	for _, key := range sorted_keys {
		if args[key] != "" {
			params += fmt.Sprintf("%s=%s&", key, args[key])
		}
	}
	params = params[:len(params)-1]
	base += url.QueryEscape(params)

	// Since we're only adding two keys, it's easier
	// and more space-efficient to just delete them
	// them copy the whole map
	delete(args, "oauth_consumer_key")
	delete(args, "method")

	// Have the full string, now hash
	hash := hmac.New(sha1.New, []byte(secret))
	hash.Write([]byte(base))
	sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Add oauth_signature as one of the args
	args["oauth_signature"] = fmt.Sprintf("%s", sha)
}

func (request *Request) composeURL() string {
	args := request.Args

	args["oauth_consumer_key"] = request.ApiKey
	args["method"] = request.Method

	s := endpoint + "?" + encodeQuery(args)
	return s
}

func (request *Request) Execute(http_method string, secret string) (res string, ret error) {
	if request.ApiKey == "" || request.Method == "" {
		return "", Error("Need both API key and method")
	}

	var call_err error
	var response *Response

	switch http_method {
	case http.MethodPost:
		request.sign(http_method, endpoint, secret)
		// TODO fixit
		request.Args["method"] = request.Method
		request.Args["oauth_consumer_key"] = request.ApiKey
		s := encodeQuery(request.Args)
		postRequest, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(s))
		if err != nil {
			return "", err
		}
		postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		response, call_err = sendPost(postRequest)
		//TODO fixit
		delete(request.Args, "method")
		delete(request.Args, "oauth_consumer_key")
	case http.MethodGet:
		request.sign(http_method, endpoint, secret)
		s := request.composeURL()

		var res *http.Response
		res, call_err = http.Get(s)
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)
		if err := xml.Unmarshal(body, response); err != nil {
			return "", err
		}
	default:
		return "", errors.New("Unsupported HTTP method")
	}
	if err := checkError(call_err, response); err != nil {
		return "", err
	}
	return response.Payload, nil
}

func checkError(err error, response *Response) error {
	if response.Error != nil {
		return errors.New(response.Error.Code + ": " + response.Error.Message)
	}
	return err
}

func encodeQuery(args map[string]string) string {
	i := 0
	s := bytes.NewBuffer(nil)
	for k, v := range args {
		if i != 0 {
			s.WriteString("&")
		}
		i++
		s.WriteString(k + "=" + url.QueryEscape(v))
	}
	return s.String()
}

func (request *Request) buildPost(url_ string, photopath string, filetype string) (*http.Request, error) {
	real_url, _ := url.Parse(url_)

	f, err := os.Open(photopath)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	f_size := stat.Size()

	request.Args["oauth_consumer_key"] = request.ApiKey

	boundary, end := "----###---###--flickr-go-rules", "\r\n"

	// Build out all of POST body sans file
	header := bytes.NewBuffer(nil)
	for k, v := range request.Args {
		header.WriteString("--" + boundary + end)
		header.WriteString("Content-Disposition: form-data; name=\"" + k + "\"" + end + end)
		header.WriteString(v + end)
	}
	header.WriteString("--" + boundary + end)
	header.WriteString("Content-Disposition: form-data; name=\"photo\"; filename=\"" + filepath.Base(photopath) + "\"" + end)
	header.WriteString("Content-Type: " + filetype + end + end)

	footer := bytes.NewBufferString(end + "--" + boundary + "--" + end)

	body_len := int64(header.Len()) + int64(footer.Len()) + f_size

	r, w := io.Pipe()
	go func() {
		pieces := []io.Reader{header, f, footer}

		for _, k := range pieces {
			_, err = io.Copy(w, k)
			if err != nil {
				w.CloseWithError(nil)
				return
			}
		}
		f.Close()
		w.Close()
	}()

	http_header := make(http.Header)
	http_header.Add("Content-Type", "multipart/form-data; boundary="+boundary)

	postRequest := &http.Request{
		Method:        http.MethodPost,
		URL:           real_url,
		Header:        http_header,
		Body:          r,
		ContentLength: body_len,
	}
	return postRequest, nil
}

func (request *Request) UploadJpeg(secret string, photopath string) (photoid string, err error) {
	return request.Upload(secret, photopath, image_jpeg)
}

func (request *Request) Upload(secret string, photopath string, filetype string) (result string, err error) {
	request.sign(http.MethodPost, uploadEndpoint, secret)
	postRequest, err := request.buildPost(uploadEndpoint, photopath, filetype)
	if err != nil {
		return "", err
	}
	response, err := sendPost(postRequest)
	if err := checkError(err, response); err != nil {
		return "", err
	}
	var photoid string
	xml.Unmarshal([]byte(response.Payload), &photoid)
	return photoid, nil
}

func (request *Request) Replace(filename string, filetype string) (response *Response, err error) {
	postRequest, err := request.buildPost(replaceEndpoint, filename, filetype)
	if err != nil {
		return nil, err
	}
	return sendPost(postRequest)
}

func sendPost(postRequest *http.Request) (response *Response, err error) {
	// Create and use TCP connection (lifted mostly wholesale from http.send)
	client := http.DefaultClient
	resp, err := client.Do(postRequest)

	if err != nil {
		return nil, err
	}
	rawBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var r Response
	err = xml.Unmarshal(rawBody, &r)

	return &r, err
}
