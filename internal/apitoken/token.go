package apitoken

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateToken(length int) (string, error) {
    if length <= 0 {
        return "", fmt.Errorf("token length must be a positive integer")
    }

    b := make([]byte, length)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }

    return hex.EncodeToString(b), nil
}
