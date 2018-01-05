package encryption

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPasswordHash(t *testing.T) {
	salt := GenerateSalt()
	h1 := PasswordHash("userpassword", salt, 4)
	h2 := PasswordHash("userpassword", salt, 4)
	require.Equal(t, h1, h2)

	salt2 := GenerateSalt()
	h3 := PasswordHash("userpassword", salt2, 4)
	require.NotEqual(t, h1, h3)

	h4 := PasswordHash("username", salt, 4)
	require.NotEqual(t, h1, h4)
	require.NotEqual(t, h3, h4)

	h5 := PasswordHash("username", salt, 4)
	require.Equal(t, h4, h5)
}

func TestGeneratePassword(t *testing.T) {
	i := 0

	// password characters range
	characterRange := make([]uint8, 10, 10)
	for i := 0; i < 10; i++ {
		characterRange[i] = uint8(i + '0')
	}
	fmt.Printf("Char range: %v\n", characterRange)

	// generate 1000 passwords
	passwordSet := make(map[string]bool)
	for i < 1000 {
		pass := GeneratePassword(12, characterRange)
		require.Len(t, pass, 12, "Generated Password does not have the correct length")

		// verify all characters are within the range
		for _, c := range pass {
			if byte(c) < '0' || byte(c) > '9' {
				t.Fatal("Unexpected byte in generated password: " + string(c))
			}
		}
		passwordSet[pass] = true
		i++
	}

	// check for uniqueness (chance of duplicates should be negligible)
	require.Len(t, passwordSet, 1000, fmt.Sprintf("Found duplicate passwords in set: %v", passwordSet))
}

var blackHole interface{}

func BenchmarkPasswordHash(b *testing.B) {
	b.ReportAllocs()
	salt := GenerateSalt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blackHole = PasswordHash("weak pass", salt, 4)
	}
}

func BenchmarkPasswordGeneration(b *testing.B) {
	b.ReportAllocs()
	charRange := DefaultPasswordCharRange()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blackHole = GeneratePassword(16, charRange)
	}
}
