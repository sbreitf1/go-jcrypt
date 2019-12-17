package main

import (
	"fmt"
	"io/ioutil"

	"github.com/sbreitf1/go-jcrypt"
)

type data struct {
	UserName string `json:"username"`
	Password string `json:"password" jcrypt:"des"`
}

func main() {
	raw, err := ioutil.ReadFile("data.json")
	if err != nil {
		panic(err)
	}

	var d data
	if err := jcrypt.Unmarshal(raw, &data); err != nil {
		panic(err)
	}

	fmt.Println(data)
}
