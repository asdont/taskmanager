package security

import (
	"crypto/sha256"
	"fmt"
)

const salt = "BtwQgG6i2Ow2"

func SaltPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password+salt)))
}
