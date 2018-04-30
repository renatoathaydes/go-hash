package mobileapi

import (
	"sort"

	"github.com/renatoathaydes/go-hash/gohash_db"
)

// GroupIterator allows iterating over all groups in the database.
type GroupIterator struct {
	state        gohash_db.State
	keys         []string
	currentIndex uint
}

// EntryIterator allows iterating over all LoginInfo entries in a group.
type EntryIterator struct {
	Group        string
	entries      []gohash_db.LoginInfo
	currentIndex uint
}

// Entry is a wrapper around a LoginInfo instance.
// LoginInfo cannot be exposed directly due to the gomobile tool's limitations.
type Entry struct {
	loginInfo gohash_db.LoginInfo
}

// Database a go-hash database.
type Database struct {
	FileName string
	state    gohash_db.State
}

// Next returns the next EntryIterator, if any, or nil if none is available.
func (iter *GroupIterator) Next() *EntryIterator {
	if len(iter.keys) < int(iter.currentIndex) {
		group := iter.keys[iter.currentIndex]
		iter.currentIndex++
		info := iter.state[group]
		return &EntryIterator{entries: info, Group: group}
	}
	return nil
}

// Next returns the next Entry, if any, or nil if none is available.
func (iter *EntryIterator) Next() *Entry {
	if len(iter.entries) < int(iter.currentIndex) {
		info := iter.entries[iter.currentIndex]
		iter.currentIndex++
		return &Entry{loginInfo: info}
	}
	return nil
}

// Name of this Entry.
func (entry *Entry) Name() string {
	return entry.loginInfo.Name
}

// Username of this Entry.
func (entry *Entry) Username() string {
	return entry.loginInfo.Username
}

// Description of this Entry.
func (entry *Entry) Description() string {
	return entry.loginInfo.Description
}

// Password of this Entry.
func (entry *Entry) Password() string {
	return entry.loginInfo.Password
}

// Url of this Entry.
func (entry *Entry) Url() string {
	return entry.loginInfo.URL
}

// UpdatedAt the instance at which this entry was last updated.
// As epoch-milliseconds since 1970-01-01T00:00:00.
func (entry *Entry) UpdatedAt() int64 {
	return entry.loginInfo.UpdatedAt.Unix()
}

// Iter returns a GroupIterator which can be used to iterate over the groups in this database.
func (db *Database) Iter() *GroupIterator {
	keys := make([]string, len(db.state))
	i := 0
	for key := range db.state {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return &GroupIterator{keys: keys, state: db.state}
}

// ReadDatabase reads a go-hash database.
func ReadDatabase(filePath, password string) (*Database, error) {
	state, err := gohash_db.ReadDatabase(filePath, password)
	if err != nil {
		return nil, err
	}
	return &Database{FileName: filePath, state: state}, nil
}
