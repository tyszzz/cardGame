package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

func getMsgId() string {
	data := []byte(GetUUID())
	hash32 := murmur3.Sum32(data) // 32-bit hash
	return fmt.Sprintf("%x", hash32)
}

func GetUUID() string {
	return uuid.New().String()
}
