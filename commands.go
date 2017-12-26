package main

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type command interface {
	// run a command with the given state, within the given group, returning the group the user should manipulate after this command is run.
	run(state *State, group string, args string, reader *bufio.Reader) string
}

type lsCommand struct{}
type createCommand struct{}

var commands = map[string]command{
	"ls":     lsCommand{},
	"create": createCommand{},
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
		if len(newGroup) > 0 {
			_, ok := (*state)[newGroup]
			if !ok {
				(*state)[newGroup] = []LoginInfo{}
				return newGroup
			}
			println("Error: group already exists")
		} else {
			println("Error: please provide a name for the group")
		}
	case "entry":
		newEntry := arg
		if len(args) > 0 {
			entries, _ := (*state)[group]
			entry := createNewEntry(newEntry, reader)
			(*state)[group] = append(entries, entry)
		} else {
			println("Error: please provide a name for the group")
		}
	default:
		println("Error: cannot create " + subCmd)
	}
	return group
}

func createNewEntry(name string, reader *bufio.Reader) (result LoginInfo) {
	read := func(prompt string) string {
		print(prompt)
		a, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		return strings.TrimSpace(a)
	}

	username := read("Enter username: ")

	var URL string
	goodURL := false
	for !goodURL {
		URL = read("Enter URL: ")
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

	description := read("Enter description: ")

	var password string
	answerAccepted := false
	for !answerAccepted {
		answer := strings.ToLower(read("Generate password? [y/n]: "))
		if len(answer) == 0 || answer == "y" {
			password = generatePassword()
			fmt.Printf("Generated password for %s!\n", name)
			fmt.Printf("To copy it to the clipboard, type 'cp %s'\n", name)
			answerAccepted = true
		} else if answer == "n" {
			for !answerAccepted {
				password = read("Please enter a password (at least 4 characters): ")
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

func generatePassword() string {
	return ""
}
