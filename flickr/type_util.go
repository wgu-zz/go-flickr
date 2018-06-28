package flickr

import (
	"gopkg.in/h2non/filetype.v1/matchers"
	"gopkg.in/h2non/filetype.v1/types"
)

func IsImage(t types.Type) bool {
	for k := range matchers.Image {
		if k == t {
			return true
		}
	}
	return false
}
