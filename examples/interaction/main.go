package main

import (
	"fmt"
	"os"

	"github.com/sbreitf1/go-jcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

type data struct {
	UserName string `json:"username"`
	Password string `json:"password" jcrypt:"aes"`
}

func main() {
	var d data
	err := jcrypt.UnmarshalFromFile("data.json", &d, &jcrypt.Options{
		GetKeyHandler: enterPassphrase,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("->", d)
	fmt.Println()
}

func enterPassphrase() ([]byte, error) {
	fmt.Println("Please enter password below: (default: secret)")
	fmt.Print("> ")
	pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println()
		return nil, err
	}
	fmt.Println("******")
	return pass, nil
}
