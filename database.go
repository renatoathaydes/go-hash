package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/renatoathaydes/go-hash/encryption"
)

const DBVERSION = "GH00"

// WriteDatabase writes the encrypted database to the given filePath with the provided state and key.
func WriteDatabase(filePath, password string, data *State) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stateBytes, err := data.bytes()
	if err != nil {
		return err
	}

	salt := encryption.GenerateSalt()
	P := encryption.PasswordHash(password, salt)
	H := encryption.CheckSum(P)

	K := encryption.GenerateRandomBytes(32)
	L := encryption.GenerateRandomBytes(32)

	B1 := encryption.Encrypt(P, K[:16])
	B2 := encryption.Encrypt(P, K[16:])
	B3 := encryption.Encrypt(P, L[:16])
	B4 := encryption.Encrypt(P, L[16:])

	encryptedState, err := encryption.Encrypt(K, stateBytes)
	if err != nil {
		return err
	}

	encryptedStateLen := len(encryptedState)
	lenE = []byte(fmt.Sprintf("%4x", encryptedStateLen))

	mac := encryption.Hmac(L, append(salt, stateBytes))

	fileOffset := 0

	// version | salt | H | B1 | B2 | B3 | B4 | len(E) | E | HMAC
	for _, b = range [][]byte{[]byte(DBVERSION), salt, H, B1, B2, B3, B4, lenE, encryptedState, mac} {
		_, err = file.WriteAt(b, fileOffset)
		if err != nil {
			return err
		}
		fileOffset += len(b)	
	}
}

// ReadDatabase reads the encrypted database from the filePath, using the given key for decryption.
func ReadDatabase(filePath string, key []byte) (State, error) {
	dbError := errors.New("Corrupt database")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileOffset int64 := 0

	version := make([]byte, 4, 4)
	readCount, err := file.ReadAt(version, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 4

	if string(version) != DBVERSION {
		return nil, errors.New("Unsupported database version")
	}

	salt := make([]byte, 32, 32)
	readCount, err = file.ReadAt(salt, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 32
	
	H := make([]byte, 512, 512)
	readCount, err = file.ReadAt(H, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 512

	P := encryption.PasswordHash(password, salt)
	expectedH := encryption.CheckSum(P)
	
	if H != expectedH {
		return nil, errors.New("Wrong password or corrupt database")
	}

	B1 := make([]byte, 64, 64)
	readCount, err = file.ReadAt(B1, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 64	
	
	B2 := make([]byte, 64, 64)
	readCount, err = file.ReadAt(B2, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 64	
	
	B3 := make([]byte, 64, 64)
	readCount, err = file.ReadAt(B3, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 64	
	
	B4 := make([]byte, 64, 64)
	readCount, err = file.ReadAt(B4, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 64

	decryptedB1, err := encryption.Decrypt(P, B1)
	if err != nil {
		return nil, dbError
	}
	decryptedB2, err := encryption.Decrypt(P, B2)
	if err != nil {
		return nil, dbError
	}

	decryptedB3, err := encryption.Decrypt(P, B3)
	if err != nil {
		return nil, dbError
	}

	decryptedB4, err := encryption.Decrypt(P, B4)
	if err != nil {
		return nil, dbError
	}

	K := append(decryptedB1, decryptedB2...)
	L := append(decryptedB3, decryptedB4...)	
	
	payloadLen := make([]byte, 4, 4)
	readCount, err := file.ReadAt(payloadLen, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += 4

	payload := make([]byte, payloadLen, payloadLen)
	readCount, err := file.ReadAt(payload, fileOffset)
	if err != nil {
		return nil, dbError
	}
	fileOffset += payloadLen

	stateBytes, err := encryption.Decrypt(L, payload)
	if err != nil {
		return nil, dbError
	}

	// the rest of the file is the HMAC
	remainingLen := len(file) - fileOffset
	if remainingLen <= 0 {
		return nil, dbError
	}
	mac := make([]byte, remainingLen, remainingLen)
	_, err := file.ReadAt(hmac, fileOffset)
	if err != nil {
		return nil, dbError
	}

	expectedMac := encryption.Hmac(L, append(salt, stateBytes))
	
	if ok := encryption.VerifyHmac(expectedHMac, mac); !ok {
		return nil, dbError
	}
	
	// decryption and validation completed successfully!
	return decodeState(stateBytes)
}
