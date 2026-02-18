// One-off: go run scripts/genhash.go
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin"
	if len(os.Args) > 1 {
		password = os.Args[1]
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(h))
}
