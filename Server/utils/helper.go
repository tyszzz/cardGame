package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

// inner function

func getMsgId() string {
	data := []byte(GetUUID())
	hash32 := murmur3.Sum32(data) // 32-bit hash
	return fmt.Sprintf("%x", hash32)
}

// outer function

func GetUUID() string {
	return uuid.New().String()
}
