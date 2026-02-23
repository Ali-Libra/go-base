package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

func HandleSignalFinal(quitHandler func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-c
	quitHandler()
}

func HandleSignal(quitHandler func(ctx context.Context), ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-c
		if quitHandler != nil {
			quitHandler(ctx)
		}
		wg.Done()
	}()

	wg.Wait()
}

// 取范围随机数[min, max)
func Random(min, max int) int {
	//	rand.Seed(time.Now().Unix())
	if min >= max {
		return min
	}
	return rand.Intn(max-min) + min
}

func RandomClosed(min, max int) int {
	//	rand.Seed(time.Now().Unix())
	if min >= max {
		return min
	}
	return rand.Intn(max+1-min) + min
}

func RandomKey(keyLen uint32) string {
	var key strings.Builder
	for i := uint32(0); i < keyLen; i++ {
		key.WriteString(strconv.Itoa(RandomClosed(1, 9)))
	}
	return key.String()
}

var letter = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

func RandomLetterKey(keyLen uint32) string {
	var key strings.Builder
	for i := uint32(0); i < keyLen; i++ {
		key.WriteByte(letter[Random(0, len(letter))])
	}
	return key.String()
}

func GetIpByNetCard(name string) (net.IP, error) {
	inter, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	addrs, err := inter.Addrs()
	if err != nil {
		return nil, err
	}

	// 获取IP地址，子网掩码
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP, nil
			}
		}
	}

	return nil, errors.New(name + " no ip can get")
}

func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

func FindAvailablePort(startPort int, endPort int) int {
	for port := startPort; port < endPort; port++ {
		if IsPortAvailable(port) {
			return port
		}
	}
	return 0
}

// StructToHashFields 通用函数：将任何struct自动转换为Redis hash fields
// 根据json标签自动生成key，支持基础类型和json/struct嵌套
func StructToHashFields(v interface{}) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	val := reflect.ValueOf(v)

	// 如果是指针，获取指向的值
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 忽略未导出字段
		if !field.IsExported() {
			continue
		}

		// 获取json标签作为key
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldVal := val.Field(i)
		var strVal string

		// 根据字段类型转换为字符串
		switch fieldVal.Kind() {
		case reflect.String:
			strVal = fieldVal.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strVal = strconv.FormatInt(fieldVal.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strVal = strconv.FormatUint(fieldVal.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			strVal = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
		case reflect.Bool:
			strVal = strconv.FormatBool(fieldVal.Bool())
		case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
			// 复杂类型使用JSON序列化
			jsonBytes, err := json.Marshal(fieldVal.Interface())
			if err != nil {
				return nil, err
			}
			strVal = string(jsonBytes)
		default:
			strVal = fmt.Sprintf("%v", fieldVal.Interface())
		}

		fields[jsonTag] = strVal
	}

	return fields, nil
}

// HashFieldsToStruct 通用函数：将Redis hash数据自动反序列化到struct
// 支持从[]interface{}（HGETALL返回格式）或map[string]string
func HashFieldsToStruct(data interface{}, v interface{}) error {
	// 转换data为map[string]string
	var values map[string]string

	switch d := data.(type) {
	case []interface{}:
		// Redis HGETALL返回的格式：[key1, val1, key2, val2, ...]
		if len(d)%2 != 0 {
			return fmt.Errorf("invalid hash data length")
		}
		values = make(map[string]string, len(d)/2)
		for i := 0; i+1 < len(d); i += 2 {
			key := fmt.Sprintf("%v", d[i])
			val := fmt.Sprintf("%v", d[i+1])
			values[key] = val
		}
	case map[string]string:
		values = d
	case map[string]interface{}:
		values = make(map[string]string, len(d))
		for k, v := range d {
			values[k] = fmt.Sprintf("%v", v)
		}
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}

	// 获取目标struct的反射值
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to struct, got %v", val.Kind())
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got pointer to %v", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// 获取json标签作为key
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		strVal, ok := values[jsonTag]
		if !ok || strVal == "" {
			continue // 如果Redis中没有该字段，跳过
		}

		fieldVal := val.Field(i)

		// 根据字段类型转换
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(strVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as int64: %v", jsonTag, err)
			}
			fieldVal.SetInt(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal, err := strconv.ParseUint(strVal, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as uint64: %v", jsonTag, err)
			}
			fieldVal.SetUint(uintVal)
		case reflect.Float32, reflect.Float64:
			floatVal, err := strconv.ParseFloat(strVal, 64)
			if err != nil {
				return fmt.Errorf("failed to parse %s as float64: %v", jsonTag, err)
			}
			fieldVal.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("failed to parse %s as bool: %v", jsonTag, err)
			}
			fieldVal.SetBool(boolVal)
		case reflect.Slice, reflect.Array, reflect.Struct, reflect.Map:
			// 复杂类型使用JSON反序列化
			if err := json.Unmarshal([]byte(strVal), fieldVal.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to unmarshal %s: %v", jsonTag, err)
			}
		default:
			return fmt.Errorf("unsupported field type: %v", fieldVal.Kind())
		}
	}

	return nil
}
