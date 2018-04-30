package mobileapi

import (
	"sort"

	"github.com/renatoathaydes/go-hash/gohash_db"
)

type GroupIterator struct {
	state        gohash_db.State
	keys         []string
	currentIndex uint
}

type EntryIterator struct {
	Group        string
	entries      []gohash_db.LoginInfo
	currentIndex uint
}

type Entry struct {
	loginInfo gohash_db.LoginInfo
}

// Database is good
type Database struct {
	FileName string
	state    gohash_db.State
}

func (iter *GroupIterator) Next() *EntryIterator {
	if len(iter.keys) < int(iter.currentIndex) {
		group := iter.keys[iter.currentIndex]
		iter.currentIndex++
		info := iter.state[group]
		return &EntryIterator{entries: info, Group: group}
	}
	return nil
}

func (iter *EntryIterator) Next() *Entry {
	if len(iter.entries) < int(iter.currentIndex) {
		info := iter.entries[iter.currentIndex]
		iter.currentIndex++
		return &Entry{loginInfo: info}
	}
	return nil
}

func (entry *Entry) Name() string {
	return entry.loginInfo.Name
}

func (entry *Entry) Username() string {
	return entry.loginInfo.Username
}

func (entry *Entry) Description() string {
	return entry.loginInfo.Description
}

func (entry *Entry) Password() string {
	return entry.loginInfo.Password
}

func (entry *Entry) Url() string {
	return entry.loginInfo.URL
}

func (entry *Entry) UpdatedAt() int64 {
	return entry.loginInfo.UpdatedAt.Unix()
}

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

// ReadDatabase reads stuff
func ReadDatabase(filePath, password string) (*Database, error) {
	state, err := gohash_db.ReadDatabase(filePath, password)
	if err != nil {
		return &Database{}, err
	}
	return &Database{FileName: filePath, state: state}, nil
}
