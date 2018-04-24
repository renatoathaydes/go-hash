package main

import (
	"bufio"
	"fmt"
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
	// run a command with the given state, within the given group.
	run(state *State, group string, args string, reader *bufio.Reader)

	// help returns helpful information about how to use this command.
	help() string

	// a full explanation of how this command works.
	longHelp() string

	// the auto-completer for this command
	completer() readline.PrefixCompleterInterface
}

type helpCommand struct {
	commands map[string]command
}

type entryCommand struct {
	entries func() []string
}

type groupCommand struct {
	groups   func() []string
	groupBox *stringBox
}

type cpCommand struct {
	entries func() []string
}

type gotoCommand struct {
	entries func() []string
}

type cmpCommand struct {
	mpBox *stringBox
}

type stringBox struct {
	value string
}

// ============= CLI creation ============= //

func createCommands(state *State, groupBox *stringBox, masterPassBox *stringBox) map[string]command {
	getGroups := func() []string {
		result := make([]string, len(*state), len(*state))
		i := 0
		for gr := range *state {
			result[i] = gr
			i++
		}
		return result
	}

	getEntries := func() []string {
		entries := (*state)[groupBox.value]
		result := make([]string, len(entries), len(entries))
		for i, e := range entries {
			result[i] = e.Name
		}
		return result
	}

	var commands = map[string]command{
		"group": groupCommand{
			groups:   getGroups,
			groupBox: groupBox,
		},
		"entry": entryCommand{
			entries: getEntries,
		},
		"cp": cpCommand{
			entries: getEntries,
		},
		"goto": gotoCommand{
			entries: getEntries,
		},
		"cmp": cmpCommand{
			mpBox: masterPassBox,
		},
	}

	commands["help"] = helpCommand{
		commands: commands,
	}

	return commands
}

func createCompleter(commands map[string]command) *readline.PrefixCompleter {
	var cmdItems = make([]readline.PrefixCompleterInterface, len(commands)+2)
	i := 0
	for _, cmd := range commands {
		cmdItems[i] = cmd.completer()
		i++
	}
	cmdItems[i] = readline.PcItem("exit")
	cmdItems[i+1] = readline.PcItem("quit")

	return readline.NewPrefixCompleter(cmdItems...)
}

// ============= Commands: Short help ============= //

func (cmd helpCommand) help() string {
	return "prints this message or help about a specific command."
}

func (cmd entryCommand) help() string {
	return "manages entries within the current group."
}

func (cmd groupCommand) help() string {
	return "manages/enters groups."
}

func (cmd cpCommand) help() string {
	return "copies an entry's field to the clipboard. Fields: -p = password, -u = username."
}

func (cmd gotoCommand) help() string {
	return "goes to the URL associated with an entry and copies its password to the clipboard."
}

func (cmd cmpCommand) help() string {
	return "changes the master password."
}

// ============= Commands: Long help ============= //

const helpUsage = `
=== help command usage ===

The help command prints helpful information.

Usage:
  help [<name>]

Without a <name> argument, the help command shows general go-hash usage, 
otherwise full information about a specific command is shown.
`

const entryUsage = `
=== entry command usage ===

The entry command is used to manage entries within the current group.
To switch to a different group, use the 'group' command (type 'help group' for more information about groups).

Usage:
  entry [-option] [<name>]

Options:
  -c <name>   create an entry.
  -d <name>   delete an entry.
  -e <name>   edit an entry.
  -r <name>   rename an entry.

Without an option or a <name> argument, the entry command simply lists all entries within the current group.

Typing 'entry <name>' will either display information about the entry, or create it if the entry does not exist.

Examples:

  # list all entries in the current group
  entry

  # delete the entry called 'hello'
  entry -d hello
`
const groupUsage = `
=== group command usage ===

The group command is used to manage groups or enter a group in order to manage its entries.

Usage:
  group [-option] [<name>]

Options:
  -c <name>   create a group.
  -d <name>   delete a group.
  -r <name>   rename a group.

Without an option or a <name> argument, the group command simply lists all groups in the database.

Typing 'group <name>' will either enter the group (so that the 'entry' command will apply to entries
within the chosen group) , or create it if it does not exist.

After entering a group, the 'entry' command applies only to the entries within the entered group.
Type 'exit' to exit a group.

A group called 'default' is used if no group is entered. This group always exists but is not
shown in the prompt as other groups, allowing the user to manage entries without using groups
explicitly.

Examples:

  # list all groups
  group

  # delete a group called 'hello'
  group -d hello
`

const copyUsage = `
=== cp command usage ===

The cp command can be used to copy information about entries to the clipboard.
That allows users to easily copy/paste the information where the information is required.

Usage:
  cp [-option] <name>

Options:
  -u <name>   copy the username.
  -p <name>   copy the password.

If an option is not provided, the username associated with the chosen entry is copied.
Information is automatically removed from the clipboard after one minute.

Examples:

  # copy the username associated with the 'hello' entry
  cp hello

  # copy the password associated with the 'other' entry
  cp -p other
`

const gotoUsage = `
=== goto command usage ===

The goto command helps users login safely into websites by opening the URL associated with
an entry directly in the default browser, then copying the password to the clipboard so that
it can be pasted into the login form without waste of time.

Usage:
  goto [-option] <name>

Options:
  -n <name>   do not copy the password.

If the -n option is not used, the entry's password is copied to the clipboard automatically.

Examples:

  # go to the web page (URL) associated with the 'hello' entry
  goto hello
`

const cmpUsage = `
=== cmp command usage ===

The cmp command is used to change the master password.

No options or arguments are accepted.
`

func (cmd helpCommand) longHelp() string {
	return helpUsage
}

func (cmd entryCommand) longHelp() string {
	return entryUsage
}

func (cmd groupCommand) longHelp() string {
	return groupUsage
}

func (cmd cpCommand) longHelp() string {
	return copyUsage
}

func (cmd gotoCommand) longHelp() string {
	return gotoUsage
}

func (cmd cmpCommand) longHelp() string {
	return cmpUsage
}

// ============= Commands: Auto-completers ============= //

func (cmd helpCommand) completer() readline.PrefixCompleterInterface {
	commands := cmd.commands
	commandItems := make([]readline.PrefixCompleterInterface, len(commands), len(commands))
	i := 0
	for name := range commands {
		commandItems[i] = readline.PcItem(name)
		i++
	}
	return readline.PcItem("help", commandItems...)
}

func commandCompleter(getValues func() []string) readline.PrefixCompleterInterface {
	resolve := func(line string) []string {
		return getValues()
	}
	return readline.PcItemDynamic(resolve)
}

func (cmd entryCommand) completer() readline.PrefixCompleterInterface {
	cmp := commandCompleter(cmd.entries)
	return readline.PcItem("entry",
		cmp,
		readline.PcItem("-c"),
		readline.PcItem("-d", cmp),
		readline.PcItem("-e", cmp),
		readline.PcItem("-r", cmp))
}

func (cmd groupCommand) completer() readline.PrefixCompleterInterface {
	cmp := commandCompleter(cmd.groups)
	return readline.PcItem("group",
		cmp,
		readline.PcItem("-c"),
		readline.PcItem("-d", cmp),
		readline.PcItem("-r", cmp))
}

func (cmd cpCommand) completer() readline.PrefixCompleterInterface {
	cmp := commandCompleter(cmd.entries)
	return readline.PcItem("cp",
		cmp,
		readline.PcItem("-u", cmp),
		readline.PcItem("-p", cmp))
}

func (cmd gotoCommand) completer() readline.PrefixCompleterInterface {
	cmp := commandCompleter(cmd.entries)
	return readline.PcItem("goto",
		cmp,
		readline.PcItem("-n", cmp))
}

func (cmd cmpCommand) completer() readline.PrefixCompleterInterface {
	return readline.PcItem("cmp")
}

// ============= Commands: run implementations ============= //

func (cmd helpCommand) run(state *State, group, args string, reader *bufio.Reader) {
	commands := cmd.commands
	if args == "" {
		println("go-hash commands:\n")
		for name, cmd := range commands {
			fmt.Printf("  %-8s %s\n", name, cmd.help())
		}
		println("\nType 'exit' to exit a group or quit if you are not within a group.")
		println("To quit from anywhere, type 'quit'.")
	} else {
		cmd, exists := commands[args]
		if exists {
			println(cmd.longHelp())
		} else {
			println("Error: command does not exist.")
		}
	}
}

func (cmd entryCommand) run(state *State, group, args string, reader *bufio.Reader) {
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
	case strings.HasPrefix(args, "-"):
		println("Error: unknown option. Type 'help entry' for usage.")
		return
	default:
		entry = args
	}

	switch {
	case CreateEntry:
		createOrShowEntry(entry, state, group, reader, true)
	case DeleteEntry:
		removeEntry(entry, state, group, reader)
	case RenameEntry:
		renameEntry(entry, state, group, reader)
	case EditEntry:
		editEntry(entry, state, group, reader)

	// no option provided, the next cases list or offer to create an entry
	case len(entry) > 0:
		createOrShowEntry(entry, state, group, reader, false)
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
}

func (cmd groupCommand) run(state *State, group, args string, reader *bufio.Reader) {
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
	case strings.HasPrefix(args, "-"):
		println("Error: unknown option. Type 'help group' for usage.")
		return
	default:
		groupName = args
	}

	switch {
	case CreateGroup:
		cmd.groupBox.value = createGroup(groupName, state, group, reader)
	case DeleteGroup:
		cmd.groupBox.value = removeGroup(groupName, state, group, reader)
	case RenameGroup:
		cmd.groupBox.value = renameGroup(groupName, state, group, reader)

	// no option selected, list or offer to create group
	case len(groupName) > 0:
		_, groupExists := (*state)[groupName]
		if groupExists {
			cmd.groupBox.value = groupName
		} else {
			newGroupWanted := yesNoQuestion("Group does not exist, do you want to create it?", reader, true)
			if newGroupWanted {
				cmd.groupBox.value = createGroup(groupName, state, group, reader)
			}
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
}

func (cmd cpCommand) run(state *State, group, args string, reader *bufio.Reader) {
	CopyPassword := false
	CopyUsername := false
	entries := (*state)[group]
	var entry string
	switch {
	case strings.HasPrefix(args, "-p"):
		CopyPassword = true
		entry = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-u"):
		CopyUsername = true
		entry = strings.TrimSpace(args[2:])
	case strings.HasPrefix(args, "-"):
		println("Error: Unknown option.")
		println("Hint: valid options are: -p (password), -u (username)")
		return
	default:
		CopyUsername = true
		entry = args
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
		if !found && strings.Contains(entry, ":") {
			// split up group:entry from user input
			parts := strings.SplitN(entry, ":", 2)
			group = parts[0]
			entry = parts[1]
			entries = (*state)[group]
			entryIndex, found = findEntryIndex(&entries, entry)
		}
		if found {
			var content string
			switch {
			case CopyPassword:
				content = entries[entryIndex].Password
			case CopyUsername:
				content = entries[entryIndex].Username
			default:
				panic("Unexpected field case")
			}
			err := clipboard.WriteAll(content)
			if err != nil {
				fmt.Printf("Error: unable to copy! Reason: %s\n", err.Error())
			} else {
				go removeFromClipboardAfterDelay(content)
			}
		} else {
			fmt.Printf("Error: entry '%s' does not exist.\n", entry)
			showEntryHint()
		}
	}
}

func (cmd gotoCommand) run(state *State, group, args string, reader *bufio.Reader) {
	entry := args
	doCopyPass := true
	if strings.HasPrefix(args, "-n ") {
		entry = strings.TrimSpace(args[3:])
		doCopyPass = false
	} else if len(args) == 0 {
		println("Error: please provide the name of the entry to goto.")
		return
	}

	entries := (*state)[group]

	entryIndex, found := findEntryIndex(&entries, entry)
	if !found && strings.Contains(entry, ":") {
		// split up group:entry from user input
		parts := strings.SplitN(entry, ":", 2)
		group = parts[0]
		entry = parts[1]
		entries = (*state)[group]
		entryIndex, found = findEntryIndex(&entries, entry)
	}

	if found {
		URL := entries[entryIndex].URL
		if len(URL) == 0 {
			println("Error: entry does not have a URL to go to.")
		} else {
			go open(URL)
			if doCopyPass {
				cpCommand{}.run(state, group, "-p "+entry, reader)
			}
		}
	} else {
		fmt.Printf("Error: entry '%s' does not exist.\n", entry)
	}
}

func (cmd cmpCommand) run(state *State, group, args string, reader *bufio.Reader) {
	if len(args) > 0 {
		println("Error: the cmp command does not accept any arguments.")
	} else {
		attempts := 5
		for {
			print("Current password: ")
			pass, err := terminal.ReadPassword(int(syscall.Stdin))
			println("")
			if err != nil {
				panic(err)
			}
			if string(pass) == cmd.mpBox.value {
				cmd.mpBox.value = createPassword()
				break
			} else if attempts == 0 {
				panic("Too many failed attempts.")
			} else {
				println("Error: incorrect password. Please try again.")
			}
			attempts--
		}
	}
}

// ============= Entry helper functions ============= //

func createOrShowEntry(entry string, state *State, group string,
	reader *bufio.Reader, createOnly bool) {
	currentGroup := group
	if len(entry) > 0 {
		entries, _ := (*state)[group]
		if strings.Contains(entry, ":") {
			// split up group:entry from user input
			parts := strings.SplitN(entry, ":", 2)
			candidateGroup := parts[0]
			candidateEntry := parts[1]
			useCandidates := ask2OptionsQuestion(
				"Option 1: create or view entry '"+candidateEntry+"' in group '"+candidateGroup+"'.\n"+
					"Option 2: create or view entry '"+entry+"' in group '"+group+"'.\n\n"+
					"Which option do you prefer?", reader, "1", "2", true)
			if useCandidates {
				group = candidateGroup
				entry = candidateEntry
				var groupExists bool
				entries, groupExists = (*state)[group]
				if !groupExists {
					newGroupWanted := yesNoQuestion("Group does not exist, do you want to create it?", reader, true)
					if newGroupWanted {
						createGroup(group, state, group, reader)
					} else {
						return
					}
				}
			}
		}

		entryIndex, exists := findEntryIndex(&entries, entry)
		if exists {
			if createOnly {
				println("Error: entry already exists.")
			} else {
				println(entries[entryIndex].String())
			}
		} else {
			doCreate := createOnly ||
				yesNoQuestion("Entry does not exist. Do you want to create it?", reader, true)
			if doCreate {
				newEntry := createOrEditEntry(entry, group, currentGroup, reader, nil)
				(*state)[group] = append(entries, newEntry)
			}
		}
	} else {
		println("Error: please provide the name of the entry to be created.")
	}
}

func renameEntry(entry string, state *State, group string, reader *bufio.Reader) {
	if len(entry) > 0 {
		entries, _ := (*state)[group]

		entryIndex, found := findEntryIndex(&entries, entry)
		if !found && strings.Contains(entry, ":") {
			// split up group:entry from user input
			parts := strings.SplitN(entry, ":", 2)
			group = parts[0]
			entry = parts[1]
			entries = (*state)[group]
			entryIndex, found = findEntryIndex(&entries, entry)
		}

		if found {
			for {
				newName := read(reader, "Please enter the new entry name: ")
				if len(newName) == 0 {
					println("Error: no name provided.")
				} else if _, taken := findEntryIndex(&entries, newName); taken {
					println("Error: name alredy taken.")
				} else {
					entries[entryIndex].Name = newName
					break
				}
			}
		} else {
			println("Error: entry does not exist.")
		}
	} else {
		println("Error: please provide the name of the entry to be renamed.")
	}
}

func editEntry(entry string, state *State, group string, reader *bufio.Reader) {
	if len(entry) > 0 {
		currentGroup := group
		entries, _ := (*state)[group]

		entryIndex, found := findEntryIndex(&entries, entry)
		if !found && strings.Contains(entry, ":") {
			// split up group:entry from user input
			parts := strings.SplitN(entry, ":", 2)
			group = parts[0]
			entry = parts[1]
			entries = (*state)[group]
			entryIndex, found = findEntryIndex(&entries, entry)
		}

		if found {
			fmt.Printf("Editing entry:\n%s\n", entries[entryIndex].String())
			println("\nHint: to keep the current value for a field, don't enter a new value.\n")
			entries[entryIndex] = createOrEditEntry(entry, currentGroup, group, reader, &entries[entryIndex])
		} else {
			println("Error: entry does not exist.")
		}
	} else {
		println("Error: please provide the name of the entry to be edited.")
	}
}

func removeEntry(entryName string, state *State, group string, reader *bufio.Reader) {
	if len(entryName) == 0 {
		println("Error: please provide the name of the entry to remove.")
	} else {
		removed := removeEntryFrom(state, group, entryName)
		if !removed {
			println("Error: entry does not exist. Are you within the correct group?")
			println("Hint: To enter a group called <group-name>, type 'group group-name'.")
		}
	}
}

func createOrEditEntry(name, group, currentGroup string, reader *bufio.Reader,
	entry *LoginInfo) (result LoginInfo) {
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
		doChangePassword = yesNoQuestion("Do you want to change the password?", reader, false)
	}

	if doChangePassword {
		doGeneratePassword := yesNoQuestion("Generate password?", reader, true)
		if doGeneratePassword {
			password = generatePassword()
			fmt.Printf("Generated password for %s!\n", name)
			entryKey := name
			if currentGroup != group {
				entryKey = group + ":" + name
			}
			fmt.Printf("Hint: To copy it to the clipboard, type 'cp -p %s'.\n", entryKey)
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

func removeEntryFrom(state *State, group, entry string) bool {
	entries := (*state)[group]
	i, found := findEntryIndex(&entries, entry)
	if !found && strings.Contains(entry, ":") {
		// split up group:entry from user input
		parts := strings.SplitN(entry, ":", 2)
		group = parts[0]
		entry = parts[1]
		entries = (*state)[group]
		i, found = findEntryIndex(&entries, entry)
	}
	if found {
		(*state)[group] = append(entries[:i], entries[i+1:]...)
	}
	return found
}

// ============= Group helper functions ============= //

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
					goAhead = yesNoQuestion(fmt.Sprintf("Are you sure you want to remove all (%d) entries of the default group?",
						entriesLen), reader, false)
					if goAhead {
						(*state)[groupName] = []LoginInfo{}
					}
				} else {
					println("Warning: cannot delete the default group and there are no entries to remove.")
				}
			} else {
				if !goAhead {
					goAhead = yesNoQuestion(fmt.Sprintf("Are you sure you want to remove group '%s' and all of its (%d) entries?",
						groupName, entriesLen), reader, false)
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

// ============= Goto helper functions ============= //

// open the specified URL in the default browser of the user.
// Copied from https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func open(url string) error {
	var cmd string
	var args []string

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

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

// ============= Other helper functions ============= //

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

func read(reader *bufio.Reader, prompt string) string {
	print(prompt)
	a, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(a)
}

func yesNoQuestion(question string, reader *bufio.Reader, useYesAsDefault bool) bool {
	return ask2OptionsQuestion(question, reader, "y", "n", useYesAsDefault)
}

func ask2OptionsQuestion(question string, reader *bufio.Reader,
	yesAnswer, noAnswer string, useYesAsDefault bool) bool {
	defaultReturn := false
	defaultAnswer := noAnswer
	if useYesAsDefault {
		defaultReturn = true
		defaultAnswer = yesAnswer
	}
	prompt := question + " [" + yesAnswer + "/" + noAnswer + "] (" + defaultAnswer + "): "
	for {
		yn := strings.ToLower(read(reader, prompt))
		if len(yn) == 0 {
			return defaultReturn
		} else if yn == yesAnswer {
			return true
		} else if yn == noAnswer {
			return false
		} else {
			fmt.Printf("Please answer '%s' or '%s' (no answer means '%s')\n", yesAnswer, noAnswer, defaultAnswer)
		}
	}
}

func removeFromClipboardAfterDelay(content string) {
	time.Sleep(60 * time.Second)
	c, err := clipboard.ReadAll()
	if err == nil && c == content {
		clipboard.WriteAll("")
	}
}
