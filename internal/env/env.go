package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback

}

func GetInt(key string, fallback int64) int64{
	if val := os.Getenv(key); val != "" {
		retorno, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fallback
		}
		return retorno
	}
	return fallback
}