package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/renatoathaydes/go-hash/encryption"

	"github.com/atotto/clipboard"
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
type removeCommand struct{}
type cpCommand struct{}

var commands = map[string]command{
	"ls":    lsCommand{},
	"group": groupCommand{},
	"entry": entryCommand{},
	"rm":    removeCommand{},
	"cp":    cpCommand{},
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
	io.WriteString(w, "go-hash commands:\n\n")
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
		case 1:
			println("There is 1 group:\n")
		default:
			fmt.Printf("There are %d groups:\n\n", groupLen)
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

func (cmd removeCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	arg := args
	removeEntry := true // if false, remove group
	switch {
	case strings.HasPrefix(args, "-e"):
		arg = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-g"):
		arg = strings.TrimSpace(args[2:])
		removeEntry = false
	}

	if removeEntry {
		return rmEntry(arg, state, group, reader)
	}
	return rmGroup(arg, state, group, reader)
}

func (cmd removeCommand) help() string {
	return "removes a group or an entry of the current group."
}

func (cmd cpCommand) run(state *State, group string, args string, reader *bufio.Reader) string {
	parts := splitTrimN(args, 2)
	field := parts[0]
	entry := parts[1]
	entries := (*state)[group]
	const (
		fieldUsername = iota
		fieldPassword
		fieldUnknown
	)
	fieldCase := fieldUnknown
	switch field {
	case "-p":
		fieldCase = fieldPassword
	case "-u":
		fieldCase = fieldUsername
	default:
		println("Error: Unknown option: " + field)
		println("Hint: valid options are: -p (password), -u (username)")
		return group
	}

	showEntryHint := func() {
		if len(entries) > 0 {
			entryNames := make([]string, len(entries))
			for i, e := range entries {
				entryNames[i] = e.Name
			}
			fmt.Printf("Hint: under the current group, %s, the following entries exist: %s\n", group, strings.Join(entryNames, ", "))
		} else if len(*state) > 1 {
			println("Hint: there are no entries under the current group! " +
				"To enter a group which contains entries, use the 'group' command. " +
				"Type 'ls' to list all groups.")
		} else {
			println("Hint: there are no entries yet! You can create a new entry with the 'entry' command! " +
				"For example, try typing 'entry gmail'")
		}
	}

	if len(entry) == 0 {
		println("Error: please provide an entry name")
		showEntryHint()
	} else {
		entryIndex, found := findEntryIndex(&entries, entry)
		if found {
			var err error
			switch fieldCase {
			case fieldPassword:
				err = clipboard.WriteAll(entries[entryIndex].Password)
			case fieldUsername:
				err = clipboard.WriteAll(entries[entryIndex].Username)
			default:
				panic("Unexpected field case")
			}
			if err != nil {
				fmt.Printf("Error: unable to copy! Reason: %s\n", err.Error())
			}
		} else {
			fmt.Printf("Error: entry '%s' does not exist\n", entry)
			showEntryHint()
		}
	}
	return group
}

func (cmd cpCommand) help() string {
	return "Copies an entry's field to the clipboard. Fields: -p = password, -u = username."
}

func rmGroup(groupName string, state *State, group string, reader *bufio.Reader) string {
	if len(groupName) == 0 {
		println("Error: please provide the name of the group to remove.")
	} else {
		entries, ok := (*state)[groupName]
		entriesLen := len(entries)
		if ok {
			goAhead := entriesLen == 0 // if there are no entries, don't bother asking for confirmation
			if groupName == "default" {
				if !goAhead {
					goAhead = yesNoQuestion(fmt.Sprintf("Are you sure you want to remove all (%d) entries of the default group? [y/n]: ",
						entriesLen), reader)
					if goAhead {
						(*state)[groupName] = []LoginInfo{}
					}
				} else {
					println("Warning: cannot delete the default group and there are no entries to remove.")
				}
			} else {
				if !goAhead {
					goAhead = yesNoQuestion(fmt.Sprintf("Are you sure you want to remove group '%s' and all of its (%d) entries? [y/n]: ",
						groupName, entriesLen), reader)
					if !goAhead {
						println("Aborted!")
					}
				}
				if goAhead {
					delete(*state, groupName)
				}
			}
		} else {
			println("Error: group does not exist.")
		}
	}
	return group
}

func rmEntry(entryName string, state *State, group string, reader *bufio.Reader) string {
	if len(entryName) == 0 {
		println("Error: please provide the name of the entry to remove.")
		println("Hint: to remove a whole group, use the -g option, e.g. rm -g <group-name>.")
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
			fmt.Printf("Hint: To copy it to the clipboard, type 'cp -p %s'.\n", name)
			answerAccepted = true
		} else if answer == "n" {
			for !answerAccepted {
				print("Please enter a password (at least 4 characters): ")
				pass, err := terminal.ReadPassword(int(syscall.Stdin))
				println("")
				if err != nil {
					panic(err)
				}
				password = string(pass)
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

func generatePassword() (password string) {
	var (
		minChar uint8 = ' '
		maxChar uint8 = '~'
	)
	charRange := make([]uint8, 1+maxChar-minChar)
	for i := 0; i < len(charRange); i++ {
		charRange[i] = minChar + uint8(i)
	}

	containsChars := func(p string, min rune, max rune) bool {
		for _, c := range p {
			if min <= c && c <= max {
				return true
			}
		}
		return false
	}

	for {
		password = encryption.GeneratePassword(16, charRange)
		if containsChars(password, '0', '9') &&
			containsChars(password, 'A', 'Z') &&
			containsChars(password, 'a', 'z') {
			break
		}
	}
	println(password)
	return
}

func read(reader *bufio.Reader, prompt string) string {
	print(prompt)
	a, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(a)
}
