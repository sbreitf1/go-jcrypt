package main

import (
	"fmt"
	"os"

	"github.com/sbreitf1/go-jcrypt"
)

type data struct {
	UserName string `json:"username"`
	Password string `json:"password" jcrypt:"aes"`
}

func main() {
	printFile("data_raw.json")
	printFile("data_mode_none.json")
	printFile("data_mode_aes.json")
}

func printFile(file string) {
	fmt.Println(fmt.Sprintf("Source file %q:", file))

	var d data
	err := jcrypt.UnmarshalFromFile(file, &d, &jcrypt.Options{
		GetKeyHandler: jcrypt.StaticKey([]byte("secret")),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("->", d)
	fmt.Println()
}
