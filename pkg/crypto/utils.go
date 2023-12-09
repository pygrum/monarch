package crypto

import "crypto/rand"

func RandomBytes(n int) []byte {
	idBytes := make([]byte, n) // 16l
	_, _ = rand.Read(idBytes)
	return idBytes
}
