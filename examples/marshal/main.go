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
	d := data{"obi wan", "deathstar"}
	raw, err := jcrypt.Marshal(d, &jcrypt.Options{
		GetKeyHandler: jcrypt.StaticKey([]byte("secret")),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(raw))

	if err := ioutil.WriteFile("data.json", raw, os.ModePerm); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
