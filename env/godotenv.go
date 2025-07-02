package env

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var envMap map[string]string

func Init() bool {
	var err error
	envMap, err = godotenv.Read()
	if err != nil {
		fmt.Printf("Error loading .env file")
		return false
	}
	return true
}

func GetEnv(key string) string {
	return envMap[key]
}

func GetEnvSlice[T string | int](key string) []T {
	raw := envMap[key]
	if raw == "" {
		return []T{}
	}

	parts := strings.Split(raw, ",")
	result := make([]T, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)

		var val T
		var err error

		switch any(val).(type) {
		case string:
			val = any(p).(T)
		case int:
			i, convErr := strconv.Atoi(p)
			if convErr != nil {
				err = convErr
			} else {
				val = any(i).(T)
			}
		default:
			panic("unsupported type")
		}

		if err == nil {
			result = append(result, val)
		}
	}
	return result
}
