package internal

import (
	"errors"
	"reflect"
)

func SetString(field reflect.Value, value string) error {
	switch field.Type().Kind() {
	case reflect.String:
		field.SetString(value)
		return nil
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(&value))
			return nil
		}
	}
	return errors.New("expects string or *string type")
}

func SetInt64(field reflect.Value, value int64) error {
	switch field.Type().Kind() {
	case reflect.Int64:
		field.SetInt(value)
		return nil
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Int64 {
			field.Set(reflect.ValueOf(&value))
			return nil
		}
	}
	return errors.New("expects int64 or *int64 type")
}

func SetFloat64(field reflect.Value, value float64) error {
	switch field.Type().Kind() {
	case reflect.Float64:
		field.SetFloat(value)
		return nil
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Float64 {
			field.Set(reflect.ValueOf(&value))
			return nil
		}
	}
	return errors.New("expects float64 or *float64 type")
}

func SetBool(field reflect.Value, value bool) error {
	switch field.Type().Kind() {
	case reflect.Bool:
		field.SetBool(value)
		return nil
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Bool {
			field.Set(reflect.ValueOf(&value))
			return nil
		}
	}
	return errors.New("expects bool or *bool type")
}
