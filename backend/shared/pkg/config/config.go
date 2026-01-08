package config

import (
	"fmt"
	"os"
)

// MustGetEnv returns the value of the environment variable key.
// It panics if the key is not set.
func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return v
}

// GetEnv returns the value of the environment variable key.
// It returns fallback if the key is not set.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
