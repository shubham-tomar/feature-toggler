package utils

import (
	"bufio"
	"os"
	"strings"
)

func GetEnv(key, fallback string) string {
	filepath := os.Getenv("ENV_FILE")
	if filepath == "" {
		filepath = "./.env"
	}
	reader, err := os.Open(filepath)
	if err != nil {
		return fallback
	}
	defer reader.Close()
	
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			kv := strings.Split(line, "=")
			if kv[0] == key {
				return kv[1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fallback
	}
	return fallback
}