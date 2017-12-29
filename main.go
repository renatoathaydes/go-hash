package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

func parentDirExists(path string) bool {
	_, err := os.Stat(filepath.Dir(path))
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic("Cannot read directory path")
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir()
	}
	return false
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

	cli, err := readline.NewEx(&readline.Config{
		Prompt:            prompt(),
		AutoComplete:      createCompleter(),
		InterruptPrompt:   "^C",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

Loop:
	for {
		cli.SetPrompt(prompt())
		line, err := cli.Readline()
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
		case "help":
			usage(cli.Stdout())
		default:
			command := commands[cmd]
			if command != nil {
				group = command.run(state, group, args, reader)
				err := WriteDatabase(dbPath, userPass, state)
				if err != nil {
					println("Error writing to database: " + err.Error())
				}
			} else if len(cmd) > 0 {
				fmt.Printf("Unknown command: '%s'. Type 'help' for usage.\n", cmd)
			}
		}
	}
}

func main() {
	var userPass string
	var state State
	println("Go-Hash version " + DBVERSION)
	println("")

	var dbFilePath string

	switch len(os.Args) {
	case 1:
		dbFilePath = getGoHashFilePath()
	case 2:
		dbFilePath = os.Args[1]
		if !parentDirExists(dbFilePath) {
			panic("The provided file is under a non-existing directory. Please create the directory manually first.")
		}
		if isDir(dbFilePath) {
			panic("The path you provided is a directory. Please provide a file.")
		}
	default:
		panic("Too many arguments provided. go-hash only accepts none or one argument: the passwords file.")
	}

	dbFile, err := os.Open(dbFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			println("No database exists yet, to create one, you need to provide a strong password first.")
			println("A strong password could be a phrase you could remember easily but that is hard to guess.")
			println("To make it harder to guess, include both upper and lower-case letters, numbers and special characters like ? and @.")
			println("If you forget this password, there's no way to recover it or your data, so be careful!\n")
			userPass = createPassword()
		} else {
			panic(err)
		}
		state = make(State)
	} else {
		// the DB exists, check if the user can open it
		dbFile.Close()
		state, userPass = openDatabase(dbFilePath)
	}

	if len(state) == 0 {
		state["default"] = []LoginInfo{}
	}

	println("\nWelcome, go-hash at your service.\n")
	runCliLoop(&state, dbFilePath, userPass)
}
