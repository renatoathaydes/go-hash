package gohash_db

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LoginInfo single entry containing login information for a particular website.
type LoginInfo struct {
	Name        string
	URL         string
	Username    string
	Password    string
	Description string
	UpdatedAt   time.Time
}

// State the actual login information persisted by the database.
type State map[string][]LoginInfo

// String human-readable representation of LoginInfo.
func (info *LoginInfo) String() string {
	return fmt.Sprintf("  %s:\n    %-16s %s\n    %-16s %s\n    %-16s %s\n    %-16s %s", info.Name,
		"username:", info.Username,
		"URL:", info.URL,
		"updatedAt:", info.UpdatedAt.Format("2006-01-02 15:04:05"),
		"description:", info.Description)
}

func (info *LoginInfo) bytes() []byte {
	var result bytes.Buffer
	enc := base64.StdEncoding.EncodeToString
	result.WriteString(enc([]byte(info.Name)))
	result.WriteString(" ")
	result.WriteString(enc([]byte(info.URL)))
	result.WriteString(" ")
	result.WriteString(enc([]byte(info.Username)))
	result.WriteString(" ")
	result.WriteString(enc([]byte(info.Password)))
	result.WriteString(" ")
	result.WriteString(enc([]byte(info.Description)))
	return result.Bytes()
}

func decodeLoginInfo(info []byte) (LoginInfo, error) {
	result := LoginInfo{}
	parts := strings.SplitN(string(info), " ", 4)
	if len(parts) != 5 {
		return result, errors.New("Invalid database format")
	}
	dec := base64.StdEncoding.DecodeString
	name, err := dec(parts[0])
	if err != nil {
		return result, err
	}
	url, err := dec(parts[1])
	if err != nil {
		return result, err
	}
	username, err := dec(parts[2])
	if err != nil {
		return result, err
	}
	password, err := dec(parts[3])
	if err != nil {
		return result, err
	}
	description, err := dec(parts[4])
	if err != nil {
		return result, err
	}

	result.Name = string(name)
	result.URL = string(url)
	result.Username = string(username)
	result.Password = string(password)
	result.Description = string(description)
	return result, nil
}

// Encode the state into Go's serialization format.
func (data *State) bytes() ([]byte, error) {
	stateBuffer := bytes.Buffer{}
	gobEncoder := gob.NewEncoder(&stateBuffer)
	err := gobEncoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return stateBuffer.Bytes(), nil
}

// Decode the state from the given bytes.
func decodeState(stateBytes []byte) (State, error) {
	var data State
	stateBuffer := bytes.Buffer{}
	stateBuffer.Write(stateBytes)
	gobDecoder := gob.NewDecoder(&stateBuffer)
	err := gobDecoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
