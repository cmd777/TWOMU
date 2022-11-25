package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var pretext string

	chlog_body, err := os.ReadFile("../CHANGELOG.md")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if strings.Contains(os.Args[1], "pre") {
		pretext = "<p align='center'><img src='https://img.shields.io/badge/-This%20is%20an%20experimental%20build%20and%20may%20not%20be%20fully%20stable-orange?style=plastic'></p>"
	}

	file, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fulltext := fmt.Sprintf("%v\n\n%v\n\n%v SHA256: %x", pretext, string(chlog_body), strings.Replace(os.Args[2], "../", "", -1), hash.Sum(nil))

	errw := os.WriteFile("../CHANGELOG.md", []byte(fulltext), 0666)
	if errw != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Ok")
}
