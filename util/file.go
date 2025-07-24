package util

import (
	"encoding/json"
	"fmt"
	"os"
)

func SaveImage(data []byte, filename string) error {
	// 创建并写入文件
	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("保存图片失败: %w", err)
	}
	return nil
}

func ReadImage(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ReadConfig(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func WriteConfig(path string, config map[string]string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
