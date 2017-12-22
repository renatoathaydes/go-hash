package main

import (
	"os"
	"golang.org/x/crypto/bcrypt"
	"github.com/mitchellh/go-homedir"
	"fmt"
	"io/ioutil"
)

type ErrMessage struct {
	err     error
	message string
}

func exitIfError(errMessage *ErrMessage) {
	if errMessage.err != nil || errMessage.message != "" {
		var message string
		if errMessage.err != nil {
			message = errMessage.err.Error()
		} else {
			message = errMessage.message
		}

		println("Error: " + message)
		os.Exit(1)
	}
}

func getGoHashFilePath() string {
	home, err := homedir.Dir()
	exitIfError(&ErrMessage{err: err})
	return home + "/.go-hash"
}

func writeHash(input string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input), 12)
	exitIfError(&ErrMessage{err: err})

	filePath := getGoHashFilePath()
	file, err := os.Create(filePath)
	exitIfError(&ErrMessage{err: err})
	defer file.Close()

	_, err = file.Write(hash)
	exitIfError(&ErrMessage{err: err})
	err = file.Sync()
	exitIfError(&ErrMessage{err: err})

	fmt.Printf("Saved hash at %s\n", file.Name())
}

func checkHash(input string) {
	filePath := getGoHashFilePath()
	file, err := os.Open(filePath)
	exitIfError(&ErrMessage{err: err})
	defer file.Close()
	hash, err := ioutil.ReadAll(file)

	if err = bcrypt.CompareHashAndPassword(hash, []byte(input)); err != nil {
		println("Error: hash could not be verified")
		os.Exit(42)
	} else {
		println("MATCH!")
	}
}

func showUsage() {
	println("go-hash is a utility to hash content and later verify it.\n" +
		"Options:\n" +
		"  write <content>\n" +
		"  check <content>")
}

func main() {
	if len(os.Args) != 3 {
		println("Error: Wrong number of arguments")
		showUsage()
		os.Exit(-1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "write":
		writeHash(os.Args[2])
	case "check":
		checkHash(os.Args[2])
	default:
		println("Unknown command: " + cmd)
		os.Exit(-1)
	}
}
