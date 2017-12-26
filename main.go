package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
}

func getGoHashFilePath() string {
	home, err := homedir.Expand("~/.go-hash")
	if err != nil {
		panic(err)
	}
	return home
}

func showUsage() {
	println("go-hash is a utility to hash content and later verify it.\n" +
		"Options:\n" +
		"  write <content>\n" +
		"  check <content>")
}

func createPassword() string {
	for i := 0; i < 10; i++ {
		print("Please enter a master password: ")
		pass, err := terminal.ReadPassword(int(syscall.Stdin))
		println("")
		if err != nil {
			panic(err)
		}
		if len(pass) > 7 {
			for i := 0; i < 3; i++ {
				print("Re-enter the password: ")
				pass2, err := terminal.ReadPassword(int(syscall.Stdin))
				println("")
				if err != nil {
					panic(err)
				}
				if len(pass2) == 0 {
					break
				}
				if bytes.Equal(pass, pass2) {
					return string(pass)
				}
				println("No match! Try again or just hit Enter to start again.")
			}
		} else {
			println("Password too short! Please use at least 8 characters")
		}
	}
	panic("Too many attempts!")
}

func openDatabase(dbFilePath string) (state State, userPass string) {
	for i := 0; i < 5; i++ {
		print("Please enter your master password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		println("")
		if err != nil {
			panic(err)
		}
		userPass = string(bytePassword)
		state, err = ReadDatabase(dbFilePath, userPass)
		if err != nil {
			println("An error occurred: " + err.Error())
		} else {
			return
		}
	}
	panic("Too many attempts!")
}

func runCliLoop(state State, userPass string) {
	prompt := "$go-hash> "
	reader := bufio.NewReader(os.Stdin)

Loop:
	for {
		print(prompt)
		cmd, err := reader.ReadString('\n')
		switch strings.TrimSpace(cmd) {
		case "write":
			err = WriteDatabase(getGoHashFilePath(), userPass, &state)
			if err != nil {
				println("Error: " + err.Error())
			}
		case "ls":
			if len(state) == 0 {
				println("Empty database")
			} else {
				for group, entries := range state {
					fmt.Printf("Group: %s (%d entries)\n", group, len(entries))
					for _, e := range entries {
						println("  - " + e.String())
					}
				}
			}
		case "exit":
			break Loop
		default:
			print("Unknown command: " + cmd)
		}
	}
}

func main() {
	var userPass string
	var state State
	println("Go-Hash version " + DBVERSION)
	println("")

	dbFilePath := getGoHashFilePath()
	dbFile, err := os.Open(dbFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			println("No database exists yet, to create one, you need to provide a strong password first.")
			println("A strong password could be a phrase you could remember easily but that is hard to guess.")
			println("Make sure to include both upper and lower-case letters, numbers and special characters like ? and @\n")
			userPass = createPassword()
		} else {
			panic(err)
		}
	} else {
		// the DB exists, check if the user can open it
		dbFile.Close()
		state, userPass = openDatabase(dbFilePath)
	}

	println("Welcome, go-hash at your service.")
	runCliLoop(state, userPass)
}
