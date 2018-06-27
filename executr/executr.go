package main

import (
	"fmt"

	"github.com/wgu/go-flickr/flickr"
)

func main() {
	requestTemplate, err := flickr.NewRequestFromCmd()
	checkErr(err)
	if requestTemplate.Dir == "" {
		request := flickr.NewRequest(requestTemplate.HttpMethod, requestTemplate.Auth, requestTemplate.AdditionalArgs, requestTemplate.Secret)
		response, err := request.Execute()
		fmt.Println(response)
		checkErr(err)
	} else {
		fmt.Println("Use uploadr instead.")
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
