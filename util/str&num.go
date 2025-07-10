package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	return uuid.New().String()
}

// Number 是支持整数和浮点数的类型约束
type Number interface {
	~int | ~int32 | ~int64 |
		~uint | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func ToString[T Number](val T) string {
	switch any(val).(type) {
	case float32, float64:
		return strconv.FormatFloat(float64(val), 'f', -1, 64)
	default:
		return strconv.FormatInt(int64(val), 10)
	}
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |
		~string
}

// SortTwo 返回 a、b 排序后的结果（小的在前）
func SortTwo[T Ordered](a, b T) (T, T) {
	if a <= b {
		return a, b
	}
	return b, a
}

func loadJsonToStruct(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}

	return nil
}

// LoadJSONToMap 加载JSON文件到map
func LoadJsonFile(filename string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := loadJsonToStruct(filename, &data)
	return data, err
}

func GenUniqueID(timestampMs int64, val int) int64 {
	if val >= 65536 || val < 0 {
		panic("val must be between 0 and 9999")
	}
	return (timestampMs << 16) | int64(val)
}
