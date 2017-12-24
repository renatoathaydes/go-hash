# go-hash

> This project is under development and is not ready for use!

Current status:

- [x] database format definition
- [ ] database format implementation
- [ ] Create database
- [ ] Load database
- [ ] CLI `ls` command
- [ ] CLI `create entry` command
- [ ] CLI `create group` command
- [ ] CLI `cp u` (copy username) command
- [ ] CLI `cp p` (copy password) command
- [ ] CLI `goto` command
- [ ] CLI `rm entry` command
- [ ] CLI `rm group` command

## Description

go-hash is a utility to store passwords and other sensitive data safely under a master password.

go-hash was designed to allow users to keep the data safely even when non-secure storage is used.

It is safe, for example, to keep your go-hash data in Dropbox or other cloud services as long as a strong master password is chosen.

The database format used by go-hash was inspired by [PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm)
and modified to benefit from the strengths of [BCrypt](https://en.wikipedia.org/wiki/Bcrypt).
The format is described in detail later in this document.

## Usage

### Make sure you install go-hash:

TBD

### Create/load a database

#### Create a new database

To create a new database in the default location (`$HOME/.go-hash`):

```
go-hash new
```

To create a new database in a custom location:

```
go-hash new /path/to/file
```

#### Load an existing database

To load the database at the default location:

```
go-hash
```

> If the default database file does not exist, you will be asked to create one.

To load a database from a custom location:

```
go-hash load /path/to/file
```

### Interact with the go-hash prompt

Once you've loaded a database file, you should enter the go-hash prompt, which looks like this:

```
$go-hash>
```

#### Listing entries

In the go-hash prompt, you can list existing passwords, for example:

```
$go-hash> ls
```

You should see the following output, indicating you have no entries in your database, but there's a group called `default`:

```
default:
  <empty>
```

#### Groups

go-hash stores information in groups. You don't need to create or enter a group, though, as a default group is used if none is specified.

To create a group called `personal`, for example:

```
$go-hash> create group personal
```

After typing that, you will enter the new group, so you should see the following prompt:

```
$go-hash:personal>
```

To exit the group, type `exit`.

While within a group, any actions you perform on entries apply under the current group. If you're not within a group, you're assumed to be under the `default` group.

#### Creating entries

To create a new entry called `github`, for example:

```
$go-hash> create entry github
```

go-hash will then ask you for the information that you want to store.

For example, when I create an entry for my GitHub account, the prompt session could look like this:

```
$go-hash> create entry github
Enter username: renatoathaydes
Enter URL: github.com
Enter description: My GitHub credentials
Generate password [y/n]: y
Generated Password for github.com!
$go-hash> 
```

If you choose to not generate a password, you can simply enter it manually, but you should probably let go-hash generate a strong password for you! You don't even need to know it, as we'll see later.

#### Using entries

To actually use go-hash to help you log in to some website, you need to copy the username/password so you can paste it in the login form.

In order to avoid phishing attacks, in which an attacker might impersonate a website and request your credentials
without you realizing you're not where you thought you were, prefer to login with the following command 
(again, using the entry called `github` as an example):

```
$go-hash> goto github
```

> Phishing attacks are the number one way that attackers use to steal credentials! You may think that you would never 
fall for a phishing attack, but evidence suggests that even the most tech-savvy of us can and often do fall for phishing
attacks without even knowing it! Attackers might just redirect you to the real website, so you think that you're safe,
while they have plenty of time to use your credentials in whatever way they see fit without you realizing it for months.

This will open the URL associated with the `github` entry using your default browser, automatically copying
the password to the clipboard, so you can just paste the password immediately into the login form, knowing that you're indeed in the correct website.

To copy your username for the entry named `github`, you can type the following:

```
$go-hash> cp u github
```

To copy the password for the entry named `github`:

```
$go-hash> cp p github
```

Or simply:

```
$go-hash> cp github
```

#### Deleting entries and groups

To delete an entry named `github`:

```
$go-hash> rm entry github
```

To delete a group named `personal`:

```
$go-hash> rm group personal
```

## Database format

go-hash uses the following database format:

```
version | salt | t | m | l | H(P) | B1 | B2 | B3 | B4 | len(E(entries)) | E(entries) | HMAC
```

where:

* `version` (4 characters) version of the database.
* `salt` (32 characters) used to hash the user's master password.
* `t` (8-digit hex) time parameter of the hash function.
* `m` (8-digit hex) memory parameter of the hash function.
* `l` (8-digit hex) password hash length.
* `P` [Argon2](https://github.com/p-h-c/phc-winner-argon2)-hash of the user's master password with a key length of `l`.
  Notice that the hash is calculated based on the user's master password and the salt.
* `H(P)` (512 characters) SHA512 checksum of the password hash.
* `B1` and `B2` are the 128-bit blocks encrypted using AES256 with `P` as the key. These blocks contain the 256 bit random
  key `K` that is used to encrypt the database entries.
* `B3` and `B4` are the 128-bit blocks encrypted using AES256 with `P` as the key. These blocks contain the 256 bit random
  key `L` that is used to calculate the HMAC.
* `len(E(entries))` the length of the encrypted database entries.
* `E(entries)` the encrypted database entries. Encryption is performed using AES256 with `K` as the key.
* `HMAC` The HMAC of the whole database up to this point, with SHA512 as the underlying hash function using `L` as the key.

This format is based on the paper by Paolo Gasti and Kasper B. Rasmussen on 
[The Security of Password Manager Database Formats](http://www.6nelweb.com/bio/papers/pwvault-ESORICS12-ext.pdf) and 
adapted from PasswordSafe's
[PWS3](https://metacpan.org/pod/distribution/Crypt-PWSafe3/lib/Crypt/PWSafe3/Databaseformat.pm) format.

## Future work

* Make passwords expirable.
* Support custom rules for generated password (to work around websites that contrain the password format).
* Create cross-platform GUIs for non-techies.
* Create browser extensions for Chrome, FireFox, MS Edge, Safari.
