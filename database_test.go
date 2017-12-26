package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func emptyDB() State {
	return State{}
}

func simpleDB() State {
	return State{
		"default": []LoginInfo{
			{Name: "google", Url: "google.com", Password: "super password"},
		},
	}
}

var knownTime, _ = time.Parse("yyyy-MM-dd", "2015-09-04")

func largeDB() State {
	return State{
		"default": []LoginInfo{
			{Name: "google", Url: "google.com", Password: "super password"},
		},
		"Personal": []LoginInfo{
			{Name: "github", Url: "github.com", Password: "easy password"},
			{Name: "facebook", Password: "other password", UpdatedAt: knownTime},
			{Name: "google", Url: "google.com", Password: "new password", UpdatedAt: knownTime},
		},
		"Work": []LoginInfo{
			{Name: "amazon", Password: "difficult password"},
			{Name: "VPN", Password: "super difficult password"},
		},
	}
}

func TestCreateAndReadDBs(t *testing.T) {
	type Ex struct {
		name string
		db   State
	}
	examples := []Ex{Ex{"SimpleDB", simpleDB()}, Ex{"EmptyDB", emptyDB()}, Ex{"LargeDB", largeDB()}}

	for _, example := range examples {
		t.Logf("Testing example: %s", example)
		tmpDbPath := os.TempDir() + "/" + example.name
		userPass := "very safe password"
		err := WriteDatabase(tmpDbPath, userPass, &example.db)
		require.NoError(t, err, "Error writing database %s", example.name)
		persistedState, err := ReadDatabase(tmpDbPath, userPass)
		require.NoError(t, err, "Error reading database: %s", example.name)
		require.Equal(t, example.db, persistedState, "The restored State (%s) is not as expected", example.name)
	}
}
