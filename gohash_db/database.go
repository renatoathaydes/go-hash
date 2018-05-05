package gohash_db

import (
	"errors"
	"log"
	"os"
	"runtime"

	"github.com/renatoathaydes/go-hash/encryption"
)

const (
	// DBVersion is the current version of the go-hash database format.
	DBVersion = "GH01"

	// PrevDBVersion is the previous version of the database format. go-hash automatically migrates
	// databases from its previous version.
	PrevDBVersion = "GH00"

	// MinDBLength      V | S  | B1 | B2 | B3 | B4 | MAC| E
	MinDBLength = 4 + 32 + 32 + 32 + 32 + 32 + 32 + 4

	// MaxDBLength the maximum allowed size of a database
	MaxDBLength = 64 * 1000 * 1024

	// Argon2Threads the fixed number of threads to use for the Argon2 Hash function.
	Argon2Threads uint8 = 4
)

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
	log.Printf("Writing salt: %x", salt)
	P := encryption.PasswordHash(password, salt, Argon2Threads)
	log.Printf("Calculated P: %x", P)

	K := encryption.GenerateRandomBytes(32)
	L := encryption.GenerateRandomBytes(32)
	log.Printf("K = %x", K)
	log.Printf("L = %x", L)

	B1, err := encryption.Encrypt(P, K[:16])
	if err != nil {
		return err
	}
	log.Printf("Writing B1 = %x", B1)
	B2, err := encryption.Encrypt(P, K[16:])
	if err != nil {
		return err
	}
	log.Printf("Writing B2 = %x", B2)
	B3, err := encryption.Encrypt(P, L[:16])
	if err != nil {
		return err
	}
	log.Printf("Writing B3 = %x", B3)
	B4, err := encryption.Encrypt(P, L[16:])
	if err != nil {
		return err
	}
	log.Printf("Writing B4 = %x", B4)

	encryptedState, err := encryption.Encrypt(K, stateBytes)
	if err != nil {
		return err
	}

	if len(encryptedState) > MaxDBLength {
		return errors.New("database too big! Cannot save it to avoid file bomb attacks. Please remove entries you don't need")
	}

	mac := encryption.Hmac(L, append(salt, stateBytes...))
	log.Printf("Generated HMAC with length %d", len(mac))

	fileOffset := 0

	// version | salt | B1 | B2 | B3 | B4 | HMAC | E
	for _, b := range [][]byte{[]byte(DBVersion), salt, B1, B2, B3, B4, mac, encryptedState} {
		_, err = file.WriteAt(b, int64(fileOffset))
		if err != nil {
			return err
		}
		fileOffset += len(b)
	}
	return nil
}

// ReadDatabase reads the encrypted database from the filePath, using the given password for decryption.
func ReadDatabase(filePath string, password string) (State, error) {
	dbError := "Corrupt database"

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// limit the size of the DB
	if fileStat.Size() < MinDBLength || fileStat.Size() > 32000000 {
		return nil, errors.New(dbError)
	}

	var fileOffset int64

	version := make([]byte, 4, 4)
	_, err = file.ReadAt(version, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 4

	var threads uint8

	switch string(version) {
	case PrevDBVersion:
		threads = uint8(runtime.NumCPU())
		log.Printf("Reading old version of database, threads param set to %d", threads)
	case DBVersion:
		threads = Argon2Threads
		log.Printf("Database version: %s", DBVersion)
	default:
		return nil, errors.New("Unsupported database version")
	}

	log.Println("Reading salt")
	salt := make([]byte, 32, 32)
	_, err = file.ReadAt(salt, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 32
	log.Println("Salt read successfully, calculating P.")

	P := encryption.PasswordHash(password, salt, threads)
	log.Printf("Calculated P, reading Bs. P = %x", P)

	B1 := make([]byte, 32, 32)
	_, err = file.ReadAt(B1, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 32
	log.Printf("Read B1: %x", B1)

	B2 := make([]byte, 32, 32)
	_, err = file.ReadAt(B2, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 32
	log.Printf("Read B2: %x", B2)

	B3 := make([]byte, 32, 32)
	_, err = file.ReadAt(B3, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 32
	log.Printf("Read B3: %x", B3)

	B4 := make([]byte, 32, 32)
	_, err = file.ReadAt(B4, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 32
	log.Printf("Read B4: %x", B4)

	decryptedB1, err := encryption.Decrypt(P, B1)
	if err != nil {
		return nil, err
	}
	log.Println("Decrypted B1")
	decryptedB2, err := encryption.Decrypt(P, B2)
	if err != nil {
		return nil, err
	}
	log.Println("Decrypted B2")

	decryptedB3, err := encryption.Decrypt(P, B3)
	if err != nil {
		return nil, err
	}
	log.Println("Decrypted B3")

	decryptedB4, err := encryption.Decrypt(P, B4)
	if err != nil {
		return nil, err
	}
	log.Println("Decrypted B4")

	K := append(decryptedB1, decryptedB2...)
	L := append(decryptedB3, decryptedB4...)

	log.Printf("Got K=%x", K)
	log.Printf("Got L=%x", L)
	log.Printf("Reading HMAC")

	mac := make([]byte, 64, 64)
	_, err = file.ReadAt(mac, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += 64

	plen := fileStat.Size() - fileOffset

	if plen > MaxDBLength {
		return nil, errors.New(dbError)
	}

	log.Printf("Reading encrypted payload with len = %d", plen)
	payload := make([]byte, plen, plen)
	_, err = file.ReadAt(payload, fileOffset)
	if err != nil {
		return nil, err
	}
	fileOffset += plen

	log.Printf("Decrypting payload")
	stateBytes, err := encryption.Decrypt(K, payload)
	if err != nil {
		return nil, errors.New(dbError)
	}

	expectedMac := encryption.Hmac(L, append(salt, stateBytes...))

	log.Printf("Verifying HMAC")
	if ok := encryption.VerifyHmac(expectedMac, mac); !ok {
		return nil, errors.New("incorrect password or corrupt database")
	}
	log.Printf("Database read successfully")

	// decryption and validation completed successfully!
	data, err := decodeState(stateBytes)

	if err == nil {
		entryCount := 0
		for _, entries := range data {
			entryCount += len(entries)
		}
		log.Printf("Decoded database, found %d groups, containing %d entries",
			len(data), entryCount)
	}
	return data, err
}
