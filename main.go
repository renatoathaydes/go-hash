package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/chzyer/readline"
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

func splitTrimN(text string, max int) []string {
	result := make([]string, max)
	parts := strings.SplitN(text, " ", max)
	for i, c := range parts {
		result[i] = strings.TrimSpace(c)
	}
	return result
}

func runCliLoop(state *State, dbPath string, userPass string) {
	group := "default"
	reader := bufio.NewReader(os.Stdin)
	prompt := func() string {
		var modifier string
		if len(group) > 0 && group != "default" {
			modifier = ":" + group
		}
		return fmt.Sprintf("\033[31mgo-hash%s»\033[0m ", modifier)
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:            prompt(),
		AutoComplete:      createCompleter(),
		InterruptPrompt:   "^C",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

Loop:
	for {
		l.SetPrompt(prompt())
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break Loop
			} else {
				continue
			}
		} else if err == io.EOF {
			break Loop
		}

		parts := splitTrimN(line, 2)
		cmd := parts[0]
		args := parts[1]

		switch cmd {
		case "exit":
			if group != "default" {
				group = "default"
			} else {
				break Loop
			}
		default:
			command := commands[cmd]
			if command != nil {
				group = command.run(state, group, args, reader)
				err := WriteDatabase(dbPath, userPass, state)
				if err != nil {
					println("Error writing to database: " + err.Error())
				}
			} else if len(cmd) > 0 {
				println("Unknown command: " + cmd)
			}
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
	runCliLoop(&state, dbFilePath, userPass)
}
