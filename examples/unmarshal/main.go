package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sbreitf1/go-jcrypt"
)

type data struct {
	UserName string `json:"username"`
	Password string `json:"password" jcrypt:"aes"`
}

func main() {
	raw, err := ioutil.ReadFile("data_mode_aes.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var d data
	if err := jcrypt.Unmarshal(raw, &d, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(d)
}
