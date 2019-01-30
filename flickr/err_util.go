package flickr

import "fmt"

func CheckErr(err error, msg ...string) {
	if err != nil {
		for m := range msg {
			fmt.Println(m)
		}
		panic(err)
	}
}
