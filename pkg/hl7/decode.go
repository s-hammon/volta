package hl7

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const messageHeader = "MSH"

var DefaultSegDelim = byte('\r')

type Decoder struct {
	data     []byte
	segments []*segment // key is zero-based idx of segment
	savedErr error
}

func NewDecoder(data []byte) *Decoder {
	d := &Decoder{}
	d.init(data, DefaultSegDelim)
	return d
}

func (d *Decoder) init(data []byte, segDelim byte) {
	if len(data) < 8 {
		d.savedErr = fmt.Errorf("message is too short (length: %d)", len(data))
		return
	}
	d.data = data
	segs, err := FastScan(data, segDelim, data[3])
	if err != nil {
		d.savedErr = err
		return
	}
	d.segments = segs

}

func Unmarshal(data []byte, v any) error {
	d := NewDecoder(data)
	if d.savedErr != nil {
		return d.savedErr
	}
	return d.Decode(v)
}

func (d *Decoder) Decode(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return fmt.Errorf("hl7: Decode(non-pointer)")
	}
	return d.decodeValue(val.Elem(), 0)
}

func (d *Decoder) decodeValue(val reflect.Value, repeatIdx int) error {
	if val.Kind() == reflect.Slice {
		elemType := val.Type().Elem()
		for i := 0; ; i++ {
			elem := reflect.New(elemType).Elem()
			if err := d.decodeStruct(elem, i); err != nil {
				return err
			}
			if isEmptyStruct(elem) {
				break
			}
			val.Set(reflect.Append(val, elem))
		}
		return nil
	}
	return d.decodeStruct(val, repeatIdx)
}

func (d *Decoder) decodeStruct(val reflect.Value, repeatIdx int) error {
	t := val.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("hl7")
		if tag == "" || tag == "-" {
			continue
		}
		parts := strings.Split(tag, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid tag: %s", tag)
		}
		segName := parts[0]
		fieldIdx, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		valStr := d.getFieldValue(segName, fieldIdx, repeatIdx)
		target := val.Field(i)
		setFieldValue(target, valStr)
	}
	return nil
}

// n is the "nth" segment repeat
func (d *Decoder) getFieldValue(seg string, idx, rep int) string {
	if seg == messageHeader {
		switch idx {
		case 1:
			return string(d.data[3])
		default:
			idx--
		}
	}
	matches := GetSegments(d.segments, seg)
	if rep >= len(matches) {
		return ""
	}
	return matches[rep].GetField(d.data, idx)
}

func setFieldValue(fVal reflect.Value, raw string) {
	if raw == "" {
		return
	}
	switch fVal.Kind() {
	case reflect.String:
		fVal.SetString(replaceEscapes(raw))
		return
	case reflect.Struct:
		var comps []string
		switch {
		case strings.Contains(raw, "^"):
			comps = strings.Split(raw, "^")
		case strings.Contains(raw, "&"):
			comps = strings.Split(raw, "&")
		default:
			comps = []string{raw}
		}
		for i := range fVal.NumField() {
			sf := fVal.Type().Field(i)
			tag := sf.Tag.Get("hl7")
			if tag == "" {
				continue
			}
			compIdx, err := strconv.Atoi(tag)
			if err != nil || compIdx > len(comps) {
				continue
			}
			compVal := fVal.Field(i)
			setFieldValue(compVal, comps[compIdx-1])
		}
	case reflect.Slice:
		repeats := strings.Split(raw, "~")
		slice := reflect.MakeSlice(fVal.Type(), 0, len(repeats))
		for _, rep := range repeats {
			elem := reflect.New(fVal.Type().Elem()).Elem()
			setFieldValue(elem, rep)
			slice = reflect.Append(slice, elem)
		}
		fVal.Set(slice)
	default:
		panic("unsupported field kind: " + fVal.Kind().String())
	}
}

func isEmptyStruct(v reflect.Value) bool {
	for i := range v.NumField() {
		if !reflect.DeepEqual(v.Field(i).Interface(), reflect.Zero(v.Field(i).Type()).Interface()) {
			return false
		}
	}
	return true
}
