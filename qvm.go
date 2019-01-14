// Commands specific to QubeOS virtual machines.
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type appvmCommand struct {
	gotoCommand
	dispvm bool
}

func (this appvmCommand) run(state *State, group, args string, reader *bufio.Reader) {
	// implementation based on gotoCommand.run()
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
			passwd := ""
			if doCopyPass {
				passwd = entries[entryIndex].Password
			}
			go this.open(group, passwd, URL)
		}
	} else {
		fmt.Printf("Error: entry '%s' does not exist.\n", entry)
	}

}

func (this appvmCommand) open(group, passwd, url string) error {
	// based on open()
	var args []string
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// https://www.qubes-os.org/doc/tips-and-tricks/#opening-links-in-your-preferred-appvm
	command := "/usr/lib/qubes/qrexec-client-vm"
	if this.dispvm {
		args = append(args, "$dispvm") // qubes syntax to open disposable vm
	} else {
		args = append(args, group) // name of appvm, same name as group
	}
	args = append(args, "qpass.ClipOpenURL")
	//args = append(args, url)

	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Println("HERE", passwd)
	io.WriteString(stdin, passwd)
	io.WriteString(stdin, "\n")
	io.WriteString(stdin, url)
	io.WriteString(stdin, "\n")
	stdin.Close()
	return err
}
