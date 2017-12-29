# go-hash

> This project is under development and is not ready for use!

Current status:

- [x] database format definition
- [x] database format implementation
- [x] Create database
- [ ] Load custom database
- [ ] Generate random password
- [ ] Clear clipboard some time after copying content to it
- [x] CLI `ls` command
- [x] CLI `entry` command
- [x] CLI `group` command
- [x] CLI `cp` (copy) command
- [x] CLI `rm` (remove) command
- [ ] CLI `goto` command

## Description

go-hash is a utility to store passwords and other sensitive data safely under a master password.

go-hash was designed to allow users to keep the data safely even when non-secure storage is used.

It is safe, for example, to keep your go-hash data in Dropbox or other cloud services as long as a strong master password is chosen.

The database format used by go-hash was inspired by [PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm)
and modified to benefit from the strengths of [Argon2](https://github.com/p-h-c/phc-winner-argon2).
The format is described in detail later in this document.

## Usage

### Make sure you install go-hash:

TBD

### Create/load a database

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

Once you've loaded a database file, you will be prompted to enter a master password for the database:

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

Type `help` to see how to use go-hash.

#### Listing entries

In the go-hash prompt, you can list existing groups with the `ls` command:

```
go-hash» ls
```

You should see the following output, indicating you have no entries in your database (but there's a group called `default`):

```
default (empty)
```

To list the entries of a particular group, say `personal`, just give the name of the group as an argument to `ls`:

```
go-hash» ls personal
```

To see a particular entry within the current group, use the `entry` command:

```
go-hash» entry github
```

> An entry's password is never displayed. go-hash only allows you to copy the password to the clipboard, as explained later.

#### Groups

go-hash stores information in groups. You don't need to create or enter a group, though, as a default group is used if none is specified.

But if you want to create a group called `personal`, for example:

```
go-hash» group personal
```

After typing that, you will enter the new group automatically, so you should see the following prompt:

```
go-hash:personal»
```

To exit a group, type `exit`. If you're not within a group, typing `exit` will quit go-hash.

While within a group, any actions you perform on entries apply under the current group.

Notice that to enter a group that already exists, say the `personal` group we've just created, you just need to type `group personal` again:

```
go-hash» group personal
go-hash:personal»
```

#### Creating entries

To create a new entry called `github`, for example, use the `entry` command:

```
go-hash» entry github
```

go-hash will then ask you for the information that you want to store.

> As we saw above, if the entry already exists, go-hash will display information about the entry rather than create it.

For example, when I create an entry for my GitHub account, the prompt session could look like this:

```
go-hash» entry github
Entry does not exist, do you want to create it? [y/n]: y
Enter username: renatoathaydes
Enter URL: github.com
Enter description: My GitHub credentials
Generate password? [y/n]: y
Generated password for github!
To copy it to the clipboard, type 'cp github'.
go-hash»
```

If you choose to not generate a password, you can simply enter it manually, but you should probably let go-hash generate a strong password for you! You don't even need to know your passwords (except for the go-hash master password, of course) when you use go-hash.

#### Using entries

To actually use go-hash to help you log in to some website, you need to copy the username/password so you can paste it in the login form.

In order to avoid phishing attacks, in which an attacker might impersonate a website and request your credentials
without you realizing you're not where you thought you were, prefer to login with the following command 
(again, using the entry called `github` as an example):

```
go-hash» goto github
```

> Phishing attacks are the number one way that attackers use to steal credentials! You may think that you would never 
fall for a phishing attack, but evidence suggests that even the most tech-savvy of us can and often do fall for phishing
attacks without even knowing it! Attackers might just redirect you to the real website, so you think that you're safe,
while they have plenty of time to use your credentials in whatever way they see fit without you realizing it for months.

This will open the URL associated with the `github` entry using your default browser, automatically copying
the password to the clipboard, so you can just paste the password immediately into the login form, knowing that you're indeed in the correct website.

To copy your username for the entry named `github` (within the current group), you can type the following:

```
go-hash» cp -u github
```

To copy the password for the entry named `github`:

```
go-hash» cp -p github
```

Or simply:

```
go-hash» cp github
```

#### Deleting entries and groups

To delete an entry named `github` (within the current group):

```
go-hash» rm -e github
```

Or simply:

```
go-hash» rm github
```

To delete a group named `personal`:

```
go-hash» rm -g personal
```

## Database format

go-hash uses the following database format:

```
version | salt | H | B1 | B2 | B3 | B4 | len(E) | E | HMAC
```

where:

* `version` (4 bytes) version of the database.
* `salt` (32 bytes) used to hash the user's master password.
* `P` [Argon2](https://github.com/p-h-c/phc-winner-argon2)-hash of the user's master password with a key length    of 32. Notice that the hash is calculated based on the user's master password and the salt.
* `H` (512 bytes) SHA512 checksum of the password hash `P`.
* `K` (32 bytes) random key used to encrypt the database entries.
* `L` (32 bytes) random key used to calculate the HMAC of the database.
* `B1` (64 bytes) the least-significant half of the `K` key after encryption with `P` used as key.
* `B2` (64 bytes) the most-significant half of the `K` key after encryption with `P` used as key.		
* `B3` (64 bytes) the least-significant half of the `L` key after encryption with `P` used as key.
* `B4` (64 bytes) the most-significant half of the `L` key after encryption with `P` used as key.		
* `len(E)` (4 bytes) the hexadecimal length, in bytes, of the encrypted database entries.
* `E` the encrypted database entries. Encryption is performed using AES256 with `K` as the key.
* `HMAC` The HMAC of the salt followed by the unencrypted, serialized version of the database entries,
   with SHA512 as the underlying hash function using `L` as the key.

The Argon2 parameters used to hash the master password are part of the database format version used, and for the current version, `GH00`, are:

* `time` = 4
* `memory` = 32 * 1024
* `key length` = 32

This format is based on the paper by Paolo Gasti and Kasper B. Rasmussen on 
[The Security of Password Manager Database Formats](http://www.6nelweb.com/bio/papers/pwvault-ESORICS12-ext.pdf) and 
adapted from PasswordSafe's
[PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm) format.

## Future work

* Make passwords expirable.
* Support custom rules for generated password (to work around websites that contrain the password format).
* Create cross-platform GUIs for non-techies.
* Create browser extensions for Chrome, FireFox, MS Edge, Safari.
