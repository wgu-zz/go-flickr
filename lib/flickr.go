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
	"strconv"
	"strings"
	"time"
)

const (
	apiEndpoint     = "https://api.flickr.com/services/rest"
	uploadEndpoint  = "https://up.flickr.com/services/upload"
	replaceEndpoint = "https://up.flickr.com/services/replace"
	imageJpeg       = "image/jpeg"
)

type Request struct {
	HttpMethod string
	Args       map[string]string
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

func NewRequest(a map[string]string) *Request {
	args := make(map[string]string)
	epoch := strconv.FormatInt(time.Now().Unix(), 10)
	args["oauth_nonce"] = epoch
	args["oauth_timestamp"] = epoch
	args["oauth_signature_method"] = "HMAC-SHA1"
	if a != nil {
		for k, v := range a {
			args[k] = v
		}
	}
	request := Request{http.MethodGet, args}
	return &request
}

func (request *Request) sign(requestUrl string, secret string) {
	args := request.Args
	delete(args, "oauth_signature")

	sorted_keys := make([]string, len(args))

	// Sort array keys
	i := 0
	for k := range args {
		sorted_keys[i] = k
		i++
	}
	sort.Strings(sorted_keys)

	// Build out ordered key-value string prefixed by secret
	base := request.HttpMethod + "&" + url.QueryEscape(requestUrl) + "&"
	var params string
	for _, key := range sorted_keys {
		params += fmt.Sprintf("%s=%s&", key, args[key])
	}
	params = params[:len(params)-1]
	base += url.QueryEscape(params)

	// Have the full string, now hash
	hash := hmac.New(sha1.New, []byte(secret))
	hash.Write([]byte(base))
	sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Add oauth_signature as one of the args
	args["oauth_signature"] = fmt.Sprintf("%s", sha)
}

func (request *Request) composeGetUrl() string {
	s := apiEndpoint + "?" + encodeQuery(request.Args)
	return s
}

func (request *Request) Execute(secret string) (res string, ret error) {
	var call_err error
	var response *Response

	switch request.HttpMethod {
	case http.MethodPost:
		request.sign(apiEndpoint, secret)
		s := encodeQuery(request.Args)
		postRequest, err := http.NewRequest(http.MethodPost, apiEndpoint, strings.NewReader(s))
		if err != nil {
			return "", err
		}
		postRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		response, call_err = sendPost(postRequest)
	case http.MethodGet:
		request.sign(apiEndpoint, secret)
		s := request.composeGetUrl()

		var res *http.Response
		res, call_err = http.Get(s)

		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
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
	if response != nil && response.Error != nil {
		return errors.New(response.Error.Code + ": " + response.Error.Message)
	}
	return err
}

func encodeQuery(args map[string]string) string {
	var b strings.Builder
	for k, v := range args {
		b.WriteString(k + "=" + url.QueryEscape(v) + "&")
	}
	return strings.TrimSuffix(b.String(), "&")
}

func (request *Request) buildPost(url_ string, photopath string, filetype string) (*http.Request, error) {
	realUrl, _ := url.Parse(url_)

	f, err := os.Open(photopath)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	f_size := stat.Size()

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

	bodyLen := int64(header.Len()) + int64(footer.Len()) + f_size

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

	httpHeader := make(http.Header)
	httpHeader.Add("Content-Type", "multipart/form-data; boundary="+boundary)

	postRequest := &http.Request{
		Method:        http.MethodPost,
		URL:           realUrl,
		Header:        httpHeader,
		Body:          r,
		ContentLength: bodyLen,
	}
	return postRequest, nil
}

func (request *Request) UploadJpeg(photopath string, secret string) (photoId string, err error) {
	return request.Upload(secret, photopath, imageJpeg)
}

func (request *Request) Upload(photopath string, filetype string, secret string) (result string, err error) {
	request.sign(uploadEndpoint, secret)
	postRequest, err := request.buildPost(uploadEndpoint, photopath, filetype)
	if err != nil {
		return "", err
	}
	response, err := sendPost(postRequest)
	if err := checkError(err, response); err != nil {
		return "", err
	}
	var photoId string
	err = xml.Unmarshal([]byte(response.Payload), &photoId)
	return photoId, err
}

//TODO Not completed yet
func (request *Request) replace(filename string, filetype string) (response *Response, err error) {
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
