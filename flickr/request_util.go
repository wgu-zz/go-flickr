package flickr

import (
	"errors"
	"flag"
	"net/http"
	"strings"
)

type RequestTemplate struct {
	HttpMethod     string
	Auth           map[string]string
	AdditionalArgs map[string]string
	Secret         string
	Dir            string
	CollectionId   string
}

func NewRequestFromCmd() (*RequestTemplate, error) {
	var httpMethod, oauth_consumer_key, oauth_token, args, secret, dir, collectionId string
	flag.StringVar(&httpMethod, "http_method", http.MethodGet, "The HTTP verb this request should use.")
	flag.StringVar(&oauth_consumer_key, "oauth_consumer_key", "", "The API Key flickr gives.")
	flag.StringVar(&oauth_token, "oauth_token", "", "The oauth token.")
	flag.StringVar(&args, "args", "", "Only for non-upload or non-replace. Arguments like flickr method, photo_id, etc. Format: \"key1=value1&key2=value2...\".")
	flag.StringVar(&secret, "secret", "", "The secret used to sign the request composed by \"api_secret&token_secret\".")
	flag.StringVar(&dir, "dir", "", "Only for upload request. Cannot be used together with `args`. The directory of photos to be uploaded.")
	flag.StringVar(&collectionId, "collection_id", "", "Optional. Only for upload request. The collection the album should be put in.")
	flag.Parse()
	if oauth_consumer_key == "" {
		return nil, errors.New("Missing oauth_consumer_key")
	}
	if oauth_token == "" {
		return nil, errors.New("Missing oauth_token")
	}
	if secret == "" {
		return nil, errors.New("Missing secret")
	}
	if args != "" && (dir != "" || collectionId != "") {
		return nil, errors.New("Either args or dir [+ collection_id] can be taken")
	}
	auth := map[string]string{
		"oauth_consumer_key": oauth_consumer_key,
		"oauth_token":        oauth_token,
	}
	additionalArgs := make(map[string]string)
	if args != "" {
		for _, s := range strings.Split(args, "&") {
			arg := strings.Split(s, "=")
			if len(arg) != 2 {
				return nil, errors.New("Wrong format of `args` " + s)
			}
			additionalArgs[arg[0]] = arg[1]
		}
	}
	return &RequestTemplate{
		httpMethod, auth, additionalArgs, secret, dir, collectionId,
	}, nil
}
