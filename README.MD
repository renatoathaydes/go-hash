# go-hash

> IMPORTANT: This project uses the latest crypto technologies to protect your data, but it comes with no warranties. If you find a weakness in the design of the database format or implementation, please create an issue immediately.

[![Build Status](https://travis-ci.org/renatoathaydes/go-hash.svg?branch=master)](https://travis-ci.org/renatoathaydes/go-hash) [![Release](https://img.shields.io/github/release/renatoathaydes/go-hash.svg)](https://github.com/renatoathaydes/go-hash/releases/latest)

Current status:

- [x] database format definition
- [x] database format implementation
- [x] Create database
- [x] Load custom database location
- [x] Generate random password
- [ ] Customize rules for generated passwords
- [x] Clear clipboard some time after copying content to it
- [x] CLI `entry` command
- [x] CLI `group` command
- [x] CLI `cp` (copy) command
- [x] CLI `cmp` (change master password) command
- [x] CLI `goto` command

## Description

go-hash is a utility to store passwords and other sensitive data safely under a master password.

go-hash was designed to allow users to keep the data safely even when non-secure storage is used.

It is safe, for example, to keep your go-hash data in Dropbox or other cloud services as long as a strong master password is chosen.

However, making sure that no one gets their hands on your data in the first place is the best way to keep it safe.
For that reason, I highly recommend using [Keybase](https://keybase.io/), a free service (similar to Dropbox, but also has chats and more),
that encrypts everything end-to-end, meaning that you can keep all your data encrypted and synchronized between your devices
without any more effort than using Dropbox.

The database format used by go-hash was inspired by [PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm)
and modified to benefit from the strengths of [Argon2](https://github.com/p-h-c/phc-winner-argon2).
The format is described in detail later in this document.

## Installing

Download the appropriate executable for your system from the [releases page](https://github.com/renatoathaydes/go-hash/releases).

> Mac users should use the `go-hash-darwin-amd64` file.

If you cannot find the correct file for your system, you will have to build from source.
See the [Building](https://github.com/renatoathaydes/go-hash#building) section for details.

Go developer? Just go get it:

```
go get -u github.com/renatoathaydes/go-hash
```

## Introduction

go-hash keeps your information safe in an encrypted file. This file is called a go-hash database.

All information in a go-hash database is organised within groups of entries.

A group has a name and may have 0 to many entries that are related somehow.

For example, you may have a `work` group for stuff releated to work, and a `personal` group for personal accounts. 

Entries contain the information you want to store securely. You can think of groups as directories, and entries as files.

Currently, the following information can be stored on each entry:

* `name` a short identifier for the entry. Used to refer to the entry with the `entry` command.
* `URL` the URL of the website this entry refers to.
* `username` your username with the given website.
* `password` your password with the given website.
* `description` a description of this entry.
* `updatedAt` last time the entry was modified.

Only `name` and `password` are mandatory.
go-hash can generate a password for you when you create the entry (or you can enter one manually if you prefer).
The `updatedAt` field is maintained automatically by go-hash.

## Usage

The first time you run `go-hash`, a new database will be created in the default location (`$HOME/.go-hash`).

```
go-hash
```

To create a new database in a custom location, just provide the path to the new file for the database:

```
go-hash /path/to/file
```

If the file already exists, go-hash will try to load it as an existing database.

### Interact with the go-hash prompt

Once you've created a database, you will be prompted to enter a master password for the database:

```
Go-Hash version GH00

No database exists yet, to create one, you need to provide a strong password first.
A strong password could be a phrase you could remember easily but that is hard to guess.
To make it harder to guess, include both upper and lower-case letters, numbers and special characters like ? and @.
If you forget this password, there's no way to recover it or your data, so be careful!

Please enter a master password:
```

Once you've done that, you should enter the go-hash prompt:

```
go-hash»
```

While on the go-hash prompt, you can use the commands explained below to interact with go-hash.

Type `help` to see information about go-hash commands, or hit Tab to see the available commands and auto-complete options.

Quit go-hash by typing `quit`.

## Commands

### group

The `group` command is used to manage groups.

To create a group called `personal`, for example, type:

```
# create or enter a group
go-hash» group personal
```

Or use the `-c` option to avoid being asked if you want to create it:

```
# create a group
go-hash» group -c personal
```

After typing that, you will enter the new group automatically, so you should see the following prompt:

```
go-hash:personal»
```

When you enter a group, you can manipulate entries within that group with the `entry` command (see the next section).

Before you enter a group, notice that you're implicitly within a `default` group, which always exists (but is not shown in the prompt).

To exit a group (actually, go back to the `default` group), type `exit`.

```
# exit a group
go-hash:personal» exit
```

To re-enter a group, just type `group personal` again. As the group already exists, this time you just enter the group instead
of being asked to create it.

You can delete a group with the `-d` option:

```
# delete a group
go-hash» group -d personal
```

To rename a group, use the `-r` option:

```
# rename a group
go-hash» group -r personal
```

You will be asked for the new name.

To list all groups, just type `group`:

```
# list all groups
go-hash» group
```

### entry

The entry command is used to show, edit, create and delete entries within the current group.

> Notice that you do not need to create a group explicitly, go-hash uses a `default` group if you do not create one.

To see all entries within the current group, just type `entry`.

```
# show all entries in the current group
go-hash» entry
```

To create an entry, give the name of the new entry, say `google`, as an argument to the `entry` command:

```
# create or show an entry called "google" within the current group
go-hash» entry google
```

If the entry does not exist, you'll be asked if you want to create it, similarly to the `group` command. To avoid the prompt,
use the `-c` option:

```
# create an entry called "google" within the current group
go-hash» entry -c google
```

You will be asked to enter the details about the entry.

Now that you've created the entry called `google`, typing `entry google` will show its details.

For example:

```
go-hash» entry google
  google:
    username:        Joe@google.com
    URL:             https://mail.google.com
    updatedAt:       2017-12-29 17:34:52
    description:     My email account.
```

> An entry's password is never displayed. go-hash only allows you to copy the password to the clipboard, as explained later.

To edit the entry, use the `-e` option:

```
# edit an entry within the current group
go-hash» entry -e google
```

You will be asked for the new details. To keep a value, just hit Enter without typing anything.

To rename an entry, use the `-r` option:

```
# rename an entry within the current group
go-hash» entry -e google
```

You will be asked for the new entry's name.

To delete an entry, use the `-d` option:

```
# delete an entry within the current group
go-hash» entry -d google
```

### goto

The safest way to login to a website is by using the `goto` command to open it in your default browser.

When you created an entry, you should have provided a `URL`... that URL is used by the `goto` command to know where to go,
so there's no way you might accidentally give your credentials to the wrong website.

Suppose you've created an entry called `google`, as explained in the `entry` command section. You can now go to the URL
associated with that entry as follows: 

```
# go to the URL associated with the "google" entry in the current group, copying the password
go-hash» goto google
```

> Phishing attacks are the number one way that attackers use to steal credentials! You may think that you would never 
fall for a phishing attack, but evidence suggests that even the most tech-savvy of us can and often do fall for phishing
attacks without even knowing it! Attackers might just redirect you to the real website, so you think that you're safe,
while they have plenty of time to use your credentials in whatever way they see fit without you realizing it for months.

The `goto` command automatically copies the entry's password into the clipboard, ready for you to paste it into the login form!

If you do not want the password to be copied automatically, use the `-n` option:

```
# go to the URL associated with the "google" entry in the current group, do not copy the password
go-hash» goto -n google
```

### cp

The `cp` (copy) command can be used to copy an entry's username and password to the clipboard, so that you can easily paste it into login forms.

To copy the password to the clipboard, use the `-p` option:

> Notice that go-hash automatically cleans up the clipboard after 1 minute, so sensitive data does not
  remain in the clipboard indefinitely.

```
# copy the password for the "google" entry in the current group
go-hash» cp -p google
```

To copy the username to the clipboard, use the `-u` option:

```
# copy the username for the "google" entry in the current group
go-hash» cp -u google
```

Or just omit any options:

```
# copy the username for the "google" entry in the current group
go-hash» cp google
```

### cmp

The `cmp` command can be used to change the opened database's master password.

Just type `cmp` and you will be prompted for the old and new passwords.

## Database format

go-hash uses the following database format:

```
version | salt | B1 | B2 | B3 | B4 | HMAC | E
```

where:

* `version` (4 bytes) version of the database ("GH00" or "GH01").
* `salt` (32 bytes) random sequence used to hash the user's master password.
* `P` (32 bytes) [Argon2](https://github.com/p-h-c/phc-winner-argon2)-hash of the user's master password.
  Notice that the hash is calculated based on the user's master password and the salt.
* `K` (32 bytes) random key used to encrypt the database entries.
* `L` (32 bytes) random key used to calculate the HMAC of the database.
* `B1` (32 bytes) the least-significant half of the `K` key after AES encryption with `P` used as key.
* `B2` (32 bytes) the most-significant half of the `K` key after AES encryption with `P` used as key.		
* `B3` (32 bytes) the least-significant half of the `L` key after AES encryption with `P` used as key.
* `B4` (32 bytes) the most-significant half of the `L` key after AES encryption with `P` used as key.		
* `HMAC` (64 bytes) The HMAC of the salt followed by the unencrypted, serialized version of the database entries,
   with SHA512 as the underlying hash function using `L` as the key.
* `E` the encrypted database entries. Encryption is performed using AES256 with `K` as the key.

The Argon2 parameters used to hash the master password are part of the database format version used, and for the current version, `GH01`, are:

* `time` = 8
* `memory` = 32 * 1024
* `key length` = 32
* `threads` = 4

For version `GH00`:

* `time` = 8
* `memory` = 32 * 1024
* `key length` = 32
* `threads` = number of CPUs (which means it was not reproducible across machines)

The encrypted length of the database proper (excluding metadat) is limited to 64 MB.

This format is based on the paper by Paolo Gasti and Kasper B. Rasmussen on 
[The Security of Password Manager Database Formats](http://www.6nelweb.com/bio/papers/pwvault-ESORICS12-ext.pdf) and 
adapted from PasswordSafe's
[PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm) format.

## Future work

* Make passwords expirable.
* Support custom rules for generated password (to work around websites that contrain the password format).
* Create cross-platform GUIs for non-techies.
* Create browser extensions for Chrome, FireFox, MS Edge, Safari.

## Building

The [releases page](https://github.com/renatoathaydes/go-hash/releases) contains executables for several platforms
but if your platform is not included or you want to build from source, just clone this repository and build it as explained below.

### Clone this repo

* Using git:

```
git clone git@github.com:renatoathaydes/go-hash.git
cd go-hash
```

* Using Go:

```
go get -u github.com/renatoathaydes/go-hash
cd $GOPATH/src/github.com/renatoathaydes/go-hash
```

### Build using make

The easiest way to build is with make. From the root directory, just run it:

```
make
```

This will get anything else required to build, then build, install and run the tests.

To run the benchmarks:

```
make bench
```

To create a local release for all of the selected targets:

```
make release
```

The local release files go in the `releases` folder.

> To see a list of all targets, in a shell that supports it (most shells), just type `make ` and hit Tab.
Otherwise, see the [Makefile](Makefile).

### Build with just go

Make is not necessary to build, it's just use for convenience.

If you don't have make or just don't want to use it, here's how to build go-hash without it:

```
# install dep if you don't have it
go get -u github.com/golang/dep/cmd/dep

# sync dependencies
dep ensure

# build or install
go build
```

Run tests with:

```
go test ./...
```
