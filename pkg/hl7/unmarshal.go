package hl7

// import (
// 	"errors"
// 	"fmt"
// 	"reflect"
// )

// var ErrInvalidStruct = errors.New("target must be a struct pointer")

// func Unmarshal(message Message, v interface{}) error {
// 	rv := reflect.ValueOf(v)
// 	if rv.Kind() != reflect.Ptr || rv.IsNil() {
// 		return ErrInvalidStruct
// 	}
// 	rv = rv.Elem()

// 	if rv.Kind() != reflect.Struct {
// 		return ErrInvalidStruct
// 	}
// 	targetType := rv.Type()

// 	for i := 0; i < rv.NumField(); i++ {
// 		f := rv.Field(i)
// 		structField := targetType.Field(i)
// 		tag := structField.Tag.Get("hl7")
// 		if !f.CanSet() || tag == "-" {
// 			continue
// 		}
// 		if v, ok := message[tag]; ok {
// 			if err := set(f, v); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func set(f reflect.Value, val interface{}) error {
// 	vv := reflect.ValueOf(val)

// 	switch vv.Kind() {
// 	case reflect.String:
// 		f.Set(vv)
// 		return nil
// 	case reflect.Map:
// 		if f.Kind() == reflect.Struct {
// 			return Unmarshal(val.(map[string]interface{}), f.Addr().Interface())
// 		}
// 		// TODO: use stringified version of map--if this doesn't work, then throw error
// 		return fmt.Errorf("hl7.Unmarshal: unhandled type: %v", vv.Kind())
// 	case reflect.Slice:
// 		if f.Kind() == reflect.Slice {
// 			vLen := vv.Len()
// 			slice := reflect.MakeSlice(f.Type(), vLen, vLen)
// 			for i := 0; i < vLen; i++ {
// 				if vv.Index(i).Kind() == reflect.Map {
// 					if err := Unmarshal(vv.Index(i).Interface().(map[string]interface{}), slice.Index(i).Addr().Interface()); err != nil {
// 						return err
// 					}
// 				} else {
// 					slice.Index(i).Set(vv.Index(i))
// 				}
// 			}
// 			f.Set(slice)
// 		}
// 		return nil
// 	default:
// 		return fmt.Errorf("hl7.Unmarshal: unhandled type: %v", vv.Kind())
// 	}
// }
