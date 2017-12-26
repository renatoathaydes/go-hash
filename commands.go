package main

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

type command interface {
	// run a command with the given state, within the given group, returning the group the user should manipulate after this command is run.
	run(state *State, group string, args string, reader *bufio.Reader) string
}

type lsCommand struct{}
type createCommand struct{}
type groupCommand struct{}

var commands = map[string]command{
	"ls":     lsCommand{},
	"create": createCommand{},
	"group":  groupCommand{},
}

func (cmd lsCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	if len(*state) == 0 {
		println("default:\n  <empty>")
	} else {
		for group, entries := range *state {
			fmt.Printf("Group: %s (%d entries)\n", group, len(entries))
			if len(entries) == 0 {
				println("  <empty>")
			} else {
				for _, e := range entries {
					println("  " + e.String())
				}
			}
		}
	}
	return group
}

func (cmd createCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	parts := splitTrimN(args, 2)
	subCmd := parts[0]
	arg := parts[1]

	switch subCmd {
	case "group":
		newGroup := arg
		ok := createNewGroup(state, newGroup)
		if ok {
			return newGroup
		}
		return group
	case "entry":
		newEntry := arg
		if len(args) > 0 {
			entries, _ := (*state)[group]
			entryExists := isEntryIn(entries, newEntry)
			if entryExists {
				println("Error: entry already exists")
			} else {
				entry := createNewEntry(newEntry, reader)
				(*state)[group] = append(entries, entry)
			}
		} else {
			println("Error: please provide a name for the group")
		}
	default:
		println("Error: cannot create " + subCmd)
	}
	return group
}

func (cmd groupCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	groupName := args
	_, groupExists := (*state)[groupName]
	if groupExists {
		return groupName
	}

	var answer bool
	for {
		yn := strings.ToLower(read(reader, "Group does not exist, do you want to create it? [y/n]: "))
		if len(yn) == 0 || yn == "y" {
			answer = true
			break
		} else if yn == "n" {
			answer = false
			break
		} else {
			println("Please answer y or n (no answer means y)")
		}
	}

	if answer {
		for {
			ok := createNewGroup(state, groupName)
			if ok {
				return groupName
			}
			groupName = read(reader, "Please enter another name for the group: ")
		}
	}
	return group
}

func createNewGroup(state *State, name string) bool {
	if len(name) > 0 {
		_, ok := (*state)[name]
		if !ok {
			(*state)[name] = []LoginInfo{}
			return true
		}
		println("Error: group already exists")
	} else {
		println("Error: please provide a name for the group")
	}
	return false
}

func read(reader *bufio.Reader, prompt string) string {
	print(prompt)
	a, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(a)
}

func createNewEntry(name string, reader *bufio.Reader) (result LoginInfo) {
	username := read(reader, "Enter username: ")

	var URL string
	goodURL := false
	for !goodURL {
		URL = read(reader, "Enter URL: ")
		if len(URL) > 0 {
			_, err := url.Parse(URL)
			if err != nil {
				println("Invalid URL, please try again.")
			} else {
				goodURL = true
			}
		} else {
			goodURL = true // empty URL is ok
		}
	}

	description := read(reader, "Enter description: ")

	var password string
	answerAccepted := false
	for !answerAccepted {
		answer := strings.ToLower(read(reader, "Generate password? [y/n]: "))
		if len(answer) == 0 || answer == "y" {
			password = generatePassword()
			fmt.Printf("Generated password for %s!\n", name)
			fmt.Printf("To copy it to the clipboard, type 'cp %s'\n", name)
			answerAccepted = true
		} else if answer == "n" {
			for !answerAccepted {
				print("Please enter a password (at least 4 characters): ")
				password, err := terminal.ReadPassword(int(syscall.Stdin))
				println("")
				if err != nil {
					panic(err)
				}
				if len(password) < 4 {
					println("Password too short, please try again!")
				} else {
					answerAccepted = true
				}
			}
		} else {
			println("Please answer y or n (no answer means y)")
		}
	}

	result.Name = name
	result.Username = username
	result.URL = URL
	result.Password = password
	result.Description = description
	result.UpdatedAt = time.Now()

	return
}

func isEntryIn(entries []LoginInfo, name string) bool {
	for _, e := range entries {
		if name == e.Name {
			return true
		}
	}
	return false
}

func generatePassword() string {
	// TODO
	return ""
}
