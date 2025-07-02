package util

import (
	"reflect"
	"strconv"
)

func StructToHash(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(v).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("redis")
		if tag == "" {
			tag = typ.Field(i).Name
		}
		result[tag] = val.Field(i).Interface()
	}
	return result
}

func HashToStruct(data map[string]string, out interface{}) error {
	val := reflect.ValueOf(out).Elem()
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("redis")
		if tag == "" {
			tag = typ.Field(i).Name
		}
		if strVal, ok := data[tag]; ok {
			field := val.Field(i)
			switch field.Kind() {
			case reflect.String:
				field.SetString(strVal)
			case reflect.Int, reflect.Int64:
				if i64, err := strconv.ParseInt(strVal, 10, 64); err == nil {
					field.SetInt(i64)
				}
			}
		}
	}
	return nil
}
