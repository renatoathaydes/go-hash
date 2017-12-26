package encryption

import (
	"testing"

	"github.com/renatoathaydes/go-hash/encryption"
	"github.com/stretchr/testify/require"
)

func TestPasswordHash(t *testing.T) {
	salt := encryption.GenerateSalt()
	h1 := encryption.PasswordHash("userpassword", salt)
	h2 := encryption.PasswordHash("userpassword", salt)
	require.Equal(t, h1, h2)

	salt2 := encryption.GenerateSalt()
	h3 := encryption.PasswordHash("userpassword", salt2)
	require.NotEqual(t, h1, h3)

	h4 := encryption.PasswordHash("username", salt)
	require.NotEqual(t, h1, h4)
	require.NotEqual(t, h3, h4)

	h5 := encryption.PasswordHash("username", salt)
	require.Equal(t, h4, h5)
}
