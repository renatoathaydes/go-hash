package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"runtime"
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

type entryCommand struct{}
type groupCommand struct{}
type cpCommand struct{}
type gotoCommand struct{}

var commands = map[string]command{
	"group": groupCommand{},
	"entry": entryCommand{},
	"cp":    cpCommand{},
	"goto":  gotoCommand{},
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

func (cmd entryCommand) run(state *State, group, args string, reader *bufio.Reader) string {
	var (
		CreateEntry bool
		DeleteEntry bool
		RenameEntry bool
		EditEntry   bool
		entry       string
	)
	switch {
	case strings.HasPrefix(args, "-c"):
		CreateEntry = true
		entry = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-d"):
		DeleteEntry = true
		entry = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-r"):
		RenameEntry = true
		entry = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-e"):
		EditEntry = true
		entry = strings.TrimSpace(args[2:])
	default:
		entry = args
	}

	switch {
	case CreateEntry:
		return createEntry(entry, state, group, reader)
	case DeleteEntry:
		return removeEntry(entry, state, group, reader)
	case RenameEntry:
		return renameEntry(entry, state, group, reader)
	case EditEntry:
		return editEntry(entry, state, group, reader)

	// no option provided, the next cases list or offer to create an entry
	case len(entry) > 0:
		entries, _ := (*state)[group]
		if entryIndex, found := findEntryIndex(&entries, entry); found {
			println(entries[entryIndex].String())
		} else {
			newEntryWanted := yesNoQuestion("Entry does not exist, do you want to create it? [y/n]: ", reader)
			if newEntryWanted {
				newEntry := createOrEditEntry(entry, reader, nil)
				(*state)[group] = append(entries, newEntry)
			}
		}
	default:
		entries := (*state)[group]
		fmt.Printf("Showing group %s:\n\n", groupDescription(group, &entries, false))
		if len(entries) > 0 {
			for _, e := range entries {
				println(e.String())
			}
		}
		println("\nHint: To show the details of a single entry, type 'entry <name>'.")
	}
	return group
}

func (cmd entryCommand) help() string {
	return "manages entries."
}

func (cmd groupCommand) run(state *State, group, args string, reader *bufio.Reader) string {
	var (
		CreateGroup bool
		DeleteGroup bool
		RenameGroup bool
		groupName   string
	)
	switch {
	case strings.HasPrefix(args, "-c"):
		CreateGroup = true
		groupName = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-d"):
		DeleteGroup = true
		groupName = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-r"):
		RenameGroup = true
		groupName = strings.TrimSpace(args[2:])
	default:
		groupName = args
	}

	switch {
	case CreateGroup:
		return createGroup(groupName, state, group, reader)
	case DeleteGroup:
		return removeGroup(groupName, state, group, reader)
	case RenameGroup:
		return renameGroup(groupName, state, group, reader)

	// no option selected, list or offer to create group
	case len(groupName) > 0:
		_, groupExists := (*state)[groupName]
		if groupExists {
			return groupName
		}

		newGroupWanted := yesNoQuestion("Group does not exist, do you want to create it? [y/n]: ", reader)
		if newGroupWanted {
			return createGroup(groupName, state, group, reader)
		}
	default:
		groupLen := len(*state)
		switch groupLen {
		case 1:
			println("There is 1 group:\n")
		default:
			fmt.Printf("There are %d groups:\n\n", groupLen)
		}
		for groupName, entries := range *state {
			fmt.Printf("  %s\n", groupDescription(groupName, &entries, true))
		}
		println("\nHint: Type 'entry' to list all entries in the current group.")
	}

	return group
}

func (cmd groupCommand) help() string {
	return "manages groups."
}

func (cmd cpCommand) run(state *State, group, args string, reader *bufio.Reader) string {
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
				"For example, try typing 'entry gmail'.")
		}
	}

	if len(entry) == 0 {
		println("Error: please provide an entry name.")
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
			fmt.Printf("Error: entry '%s' does not exist.\n", entry)
			showEntryHint()
		}
	}
	return group
}

func (cmd cpCommand) help() string {
	return "Copies an entry's field to the clipboard. Fields: -p = password, -u = username."
}

func (cmd gotoCommand) run(state *State, group, args string, reader *bufio.Reader) string {
	entryName := args
	doCopyPass := true
	if strings.HasPrefix(args, "-n ") {
		entryName = strings.TrimSpace(args[3:])
		doCopyPass = false
	} else if len(args) == 0 {
		println("Error: please provide the name of the entry to goto.")
		return group
	}

	entries := (*state)[group]
	if entryIndex, found := findEntryIndex(&entries, entryName); found {
		URL := entries[entryIndex].URL
		if len(URL) == 0 {
			println("Error: entry does not have a URL to go to.")
		} else {
			go open(URL)
			if doCopyPass {
				return cpCommand{}.run(state, group, "-p "+entryName, reader)
			}
		}
	} else {
		fmt.Printf("Error: entry '%s' does not exist.\n", entryName)
	}
	return group
}

func (cmd gotoCommand) help() string {
	return "Goes to the URL associated with an entry and copies its password to the clipboard."
}

func createEntry(entry string, state *State, group string, reader *bufio.Reader) string {
	if len(entry) > 0 {
		entries, _ := (*state)[group]
		if _, exists := findEntryIndex(&entries, entry); exists {
			println("Error: entry already exists.")
		} else {
			newEntry := createOrEditEntry(entry, reader, nil)
			(*state)[group] = append(entries, newEntry)
		}
	} else {
		println("Error: please provide the name of the entry to be created.")
	}
	return group
}

func renameEntry(entry string, state *State, group string, reader *bufio.Reader) string {
	if len(entry) > 0 {
		entries, _ := (*state)[group]
		if index, exists := findEntryIndex(&entries, entry); exists {
			for {
				newName := read(reader, "Please enter the new entry name: ")
				if len(newName) == 0 {
					println("Error: no name provided.")
				} else if _, taken := findEntryIndex(&entries, newName); taken {
					println("Error: name alredy taken.")
				} else {
					entries[index].Name = newName
					break
				}
			}
		} else {
			println("Error: entry does not exist.")
		}
	} else {
		println("Error: please provide the name of the entry to be renamed.")
	}
	return group
}

func editEntry(entry string, state *State, group string, reader *bufio.Reader) string {
	if len(entry) > 0 {
		entries, _ := (*state)[group]
		if index, exists := findEntryIndex(&entries, entry); exists {
			fmt.Printf("Editing entry:\n%s\n", entries[index].String())
			println("\nHint: to keep the current value for a field, don't enter a new value.\n")
			entries[index] = createOrEditEntry(entry, reader, &entries[index])
		} else {
			println("Error: entry does not exist.")
		}
	} else {
		println("Error: please provide the name of the entry to be edited.")
	}
	return group
}

func createGroup(name string, state *State, group string, reader *bufio.Reader) string {
	if len(name) > 0 {
		_, ok := (*state)[name]
		if !ok {
			(*state)[name] = []LoginInfo{}
			return name
		}
		println("Error: group already exists.")
	} else {
		println("Error: please provide a name for the group.")
	}
	return group
}

func renameGroup(name string, state *State, group string, reader *bufio.Reader) string {
	if len(name) > 0 {
		entries, ok := (*state)[name]
		if ok {
			var newGroupName string
			for {
				newGroupName = read(reader, "Enter a new name for the group: ")
				if len(newGroupName) > 0 {
					_, exists := (*state)[newGroupName]
					if exists {
						println("Error: name already taken.")
					} else {
						break
					}
				} else {
					println("Error: no name provided.")
				}
			}
			if name == "default" {
				(*state)["default"] = []LoginInfo{}
			} else {
				delete(*state, name)
			}
			(*state)[newGroupName] = entries
			if name == group {
				return newGroupName
			}
		} else {
			println("Error: Group does not exist.")
		}
	} else {
		println("Error: please provide the name of the group to be renamed.")
	}
	return group
}

func removeGroup(groupName string, state *State, group string, reader *bufio.Reader) string {
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
					if group == groupName {
						return "default" // exit the deleted group
					}
				}
			}
		} else {
			println("Error: group does not exist.")
		}
	}
	return group
}

func removeEntry(entryName string, state *State, group string, reader *bufio.Reader) string {
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

func groupDescription(name string, entries *[]LoginInfo, tabularFormat bool) string {
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
	template := "%-16s (%s)"
	if !tabularFormat {
		template = "%s (%s)"
	}
	return fmt.Sprintf(template, name, entriesSize)
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

func createOrEditEntry(name string, reader *bufio.Reader, entry *LoginInfo) (result LoginInfo) {
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
	doChangePassword := true
	if entry != nil {
		doChangePassword = yesNoQuestion("Do you want to change the password? [y/n]: ", reader)
	}

	if doChangePassword {
		doGeneratePassword := yesNoQuestion("Generate password? [y/n]: ", reader)
		if doGeneratePassword {
			password = generatePassword()
			fmt.Printf("Generated password for %s!\n", name)
			fmt.Printf("Hint: To copy it to the clipboard, type 'cp -p %s'.\n", name)
		} else {
			for {
				print("Please enter a password (at least 4 characters): ")
				pass, err := terminal.ReadPassword(int(syscall.Stdin))
				println("")
				if err != nil {
					panic(err)
				}
				password = string(pass)
				if len(password) < 4 {
					println("Error: Password too short, please try again!")
				} else {
					break
				}
			}
		}
	}

	if entry != nil {
		if username == "" {
			username = entry.Username
		}
		if URL == "" {
			URL = entry.URL
		}
		if password == "" {
			password = entry.Password
		}
		if description == "" {
			description = entry.Description
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
	charRange := encryption.DefaultPasswordCharRange()

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
	return
}

// open the specified URL in the default browser of the user.
// Copied from https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func read(reader *bufio.Reader, prompt string) string {
	print(prompt)
	a, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(a)
}
