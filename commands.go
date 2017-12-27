package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"golang.org/x/crypto/ssh/terminal"
)

type command interface {
	// run a command with the given state, within the given group, returning the group the user should manipulate after this command is run.
	run(state *State, group string, args string, reader *bufio.Reader) string

	// help returns helpful information about how to use this command.
	help() string
}

type lsCommand struct{}
type entryCommand struct{}
type groupCommand struct{}
type removeEntryCommand struct{}

var commands = map[string]command{
	"ls":    lsCommand{},
	"group": groupCommand{},
	"entry": entryCommand{},
	"rm":    removeEntryCommand{},
}

func createCompleter() *readline.PrefixCompleter {
	var cmdItems = make([]readline.PrefixCompleterInterface, len(commands)+1)
	i := 0
	for k := range commands {
		cmdItems[i] = readline.PcItem(k)
		i++
	}
	cmdItems[i] = readline.PcItem("exit")
	return readline.NewPrefixCompleter(cmdItems...)
}

func usage(w io.Writer) {
	io.WriteString(w, "go-hash commands:\n")
	for cmd, c := range commands {
		io.WriteString(w, fmt.Sprintf("  %-8s %s\n", cmd, c.help()))
	}
	io.WriteString(w, "\nType 'exit' to exit a group or quit if you are not within a group.\n")
	io.WriteString(w, "Type 'help' to print this message.\n")
}

func (cmd lsCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	if len(args) > 0 {
		groupName := args
		entries, ok := (*state)[groupName]
		if ok {
			fmt.Printf("Group %s:\n", groupDescription(groupName, &entries))
			if len(entries) > 0 {
				for _, e := range entries {
					println(e.String())
				}
			}
		} else {
			println("Group does not exist: " + groupName)
		}
	} else {
		groupLen := len(*state)
		switch groupLen {
		case 0:
			println("1 group:\n")
			fmt.Printf("  %s\n", groupDescription("default", &[]LoginInfo{}))
		case 1:
			println("1 group:\n")
		default:
			fmt.Printf("%d groups:\n\n", groupLen)
		}
		for groupName, entries := range *state {
			fmt.Printf("  %s\n", groupDescription(groupName, &entries))
		}
		println("\nHint: Type 'ls group-name' to see all entries in a group called <group-name>.")
	}
	return group
}

func (cmd lsCommand) help() string {
	return "shows all groups, or a group's entries if given a group name."
}

func (cmd entryCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	newEntry := args
	if len(newEntry) > 0 {
		entries, _ := (*state)[group]
		if entryIndex, found := findEntryIndex(&entries, newEntry); found {
			println(entries[entryIndex].String())
		} else {
			newEntryWanted := yesNoQuestion("Entry does not exist, do you want to create it? [y/n]: ", reader)
			if newEntryWanted {
				entry := createNewEntry(newEntry, reader)
				(*state)[group] = append(entries, entry)
			}
		}
	} else {
		println("Error: please provide a name for the entry")
	}
	return group
}

func (cmd entryCommand) help() string {
	return "shows/creates an entry."
}

func (cmd groupCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	groupName := args
	if len(groupName) > 0 {
		_, groupExists := (*state)[groupName]
		if groupExists {
			return groupName
		}

		newGroupWanted := yesNoQuestion("Group does not exist, do you want to create it? [y/n]: ", reader)
		if newGroupWanted {
			for {
				ok := createNewGroup(state, groupName)
				if ok {
					return groupName
				}
				groupName = read(reader, "Please enter another name for the group: ")
			}
		}
	} else {
		println("Error: please provide a name for the group")
	}

	return group
}

func (cmd groupCommand) help() string {
	return "enters/creates a group."
}

func (cmd removeEntryCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	entryName := args
	if len(entryName) == 0 {
		println("Error: please provide the name of the entry to remove.")
	} else {
		removed := false
		entries, ok := (*state)[group]
		if ok {
			entries, removed = removeEntryFrom(&entries, entryName)
			if removed {
				(*state)[group] = entries
			}
		}
		if !removed {
			println("Error: entry does not exist. Are you within the correct group?")
			println("Hint: To enter a group called <group-name>, type 'group group-name'.")
		}
	}
	return group
}

func (cmd removeEntryCommand) help() string {
	return "enters/creates a group."
}

func groupDescription(name string, entries *[]LoginInfo) string {
	var entriesSize string
	entriesLen := len(*entries)
	switch entriesLen {
	case 0:
		entriesSize = "empty"
	case 1:
		entriesSize = "1 entry"
	default:
		entriesSize = fmt.Sprintf("%d entries", entriesLen)
	}
	return fmt.Sprintf("%-16s (%s)", name, entriesSize)
}

func yesNoQuestion(question string, reader *bufio.Reader) bool {
	for {
		yn := strings.ToLower(read(reader, question))
		if len(yn) == 0 || yn == "y" {
			return true
		} else if yn == "n" {
			return false
		} else {
			println("Please answer y or n (no answer means y)")
		}
	}

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
			fmt.Printf("To copy it to the clipboard, type 'cp %s'.\n", name)
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

func findEntryIndex(entries *[]LoginInfo, name string) (int, bool) {
	for i, e := range *entries {
		if name == e.Name {
			return i, true
		}
	}
	return -1, false
}

func removeEntryFrom(entries *[]LoginInfo, name string) ([]LoginInfo, bool) {
	if i, found := findEntryIndex(entries, name); found {
		return append((*entries)[:i], (*entries)[i+1:]...), true
	}
	return *entries, false
}

func generatePassword() string {
	// TODO
	return ""
}
