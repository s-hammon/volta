package hl7

//
// import (
// 	"cmp"
// 	"fmt"
// 	"reflect"
// 	"slices"
// 	"strings"
// 	"sync"
// 	"unicode"
// )
//
// func Unmarshal(data []byte, v any) error {
// 	var d decodeState
// 	d.init(data)
// 	if d.savedError != nil {
// 		return d.savedError
// 	}
// 	return d.unmarshal(v)
// }
//
// type UnmarshalTypeError struct {
// 	Value  string
// 	Type   reflect.Type
// 	Offset int64
// 	Struct string
// 	Field  string
// }
//
// func (e *UnmarshalTypeError) Error() string {
// 	if e.Struct != "" || e.Field != "" {
// 		return "hl7: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
// 	}
// 	return "hl7: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
// }
//
// type InvalidUnmarshalError struct {
// 	Type reflect.Type
// }
//
// func (e *InvalidUnmarshalError) Error() string {
// 	if e.Type == nil {
// 		return "hl7: Unmarshal(nil)"
// 	}
// 	if e.Type.Kind() != reflect.Pointer {
// 		return "hl7: Unmarshal(non-pointer " + e.Type.String() + ")"
// 	}
// 	return "hl7: Unmarshal(nil " + e.Type.String() + ")"
// }
//
// type field struct {
// 	name      string
// 	nameBytes []byte
// 	tag       bool
// 	idx       []int
// 	typ       reflect.Type
// }
//
// type structFields struct {
// 	list   []field
// 	byName map[string]*field
// }
//
// var fieldCache sync.Map
//
// func typeFields(t reflect.Type) structFields {
// 	current := []field{}
// 	next := []field{{typ: t}}
//
// 	var count, nextCount map[reflect.Type]int
// 	visited := map[reflect.Type]bool{}
//
// 	var fields []field
//
// 	for len(next) > 0 {
// 		current, next = next, current[:0]
// 		count, nextCount = nextCount, map[reflect.Type]int{}
//
// 		for _, f := range current {
// 			if visited[f.typ] {
// 				continue
// 			}
// 			visited[f.typ] = true
//
// 			for i := range f.typ.NumField() {
// 				sf := f.typ.Field(i)
// 				if sf.Anonymous {
// 					t := sf.Type
// 					if t.Kind() == reflect.Pointer {
// 						t = t.Elem()
// 					}
// 					if !sf.IsExported() && t.Kind() != reflect.Struct {
// 						continue
// 					}
// 				} else if !sf.IsExported() {
// 					continue
// 				}
// 				tag := sf.Tag.Get("hl7")
// 				if tag == "-" {
// 					continue
// 				}
// 				name, _ := parseTag(tag) // may change this to return variable list of tags (e.g. ORC,OBR)
// 				if !isValidTag(name) {
// 					name = ""
// 				}
// 				idx := make([]int, len(f.idx)+1)
// 				copy(idx, f.idx)
// 				idx[len(f.idx)] = i
//
// 				ft := sf.Type
// 				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
// 					ft = ft.Elem()
// 				}
//
// 				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
// 					tagged := name != ""
// 					if name == "" {
// 						name = sf.Name
// 					}
// 					field := field{
// 						name: name,
// 						tag:  tagged,
// 						idx:  idx,
// 						typ:  ft,
// 					}
// 					field.nameBytes = []byte(field.name)
// 					fields = append(fields, field)
// 					if count[f.typ] > 1 {
// 						fields = append(fields, fields[len(fields)-1])
// 					}
// 					continue
// 				}
//
// 				nextCount[ft]++
// 				if nextCount[ft] == 1 {
// 					next = append(next, field{name: ft.Name(), idx: idx, typ: ft})
// 				}
// 			}
// 		}
// 	}
// 	slices.SortFunc(fields, func(a, b field) int {
// 		if c := strings.Compare(a.name, b.name); c != 0 {
// 			return c
// 		}
// 		if c := cmp.Compare(len(a.idx), len(b.idx)); c != 0 {
// 			return c
// 		}
// 		if a.tag != b.tag {
// 			if a.tag {
// 				return -1
// 			}
// 			return +1
// 		}
// 		return slices.Compare(a.idx, b.idx)
// 	})
// 	out := fields[:0]
// 	for advance, i := 0, 0; i < len(fields); i += advance {
// 		fi := fields[i]
// 		name := fi.name
// 		for advance = 1; i+advance < len(fields); advance++ {
// 			fj := fields[i+advance]
// 			if fj.name != name {
// 				break
// 			}
// 		}
// 		if advance == 1 {
// 			out = append(out, fi)
// 			continue
// 		}
// 		dominant, ok := dominantField(fields[i : i+advance])
// 		if ok {
// 			out = append(out, dominant)
// 		}
// 	}
// 	fields = out
// 	slices.SortFunc(fields, func(i, j field) int {
// 		return slices.Compare(i.idx, j.idx)
// 	})
//
// 	exactNameIdx := make(map[string]*field, len(fields))
// 	for i, field := range fields {
// 		exactNameIdx[field.name] = &fields[i]
// 	}
// 	return structFields{fields, exactNameIdx}
// }
//
// func dominantField(fields []field) (field, bool) {
// 	if len(fields) > 1 && len(fields[0].idx) == len(fields[1].idx) && fields[0].tag == fields[1].tag {
// 		return field{}, false
// 	}
// 	return fields[0], true
// }
//
// func isValidTag(s string) bool {
// 	if s == "" {
// 		return false
// 	}
// 	for _, c := range s {
// 		switch {
// 		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
// 		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
// 			return false
// 		}
// 	}
// 	return true
// }
//
// func cachedTypeFields(t reflect.Type) structFields {
// 	if f, ok := fieldCache.Load(t); ok {
// 		return f.(structFields)
// 	}
// 	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
// 	return f.(structFields)
// }
//
// type decodeState struct {
// 	data         []byte
// 	off          int
// 	lastCode     int
// 	scan         scanner
// 	errorContext *errorContext
// 	savedError   error
// }
//
// func (d *decodeState) readIndex() int {
// 	return d.off - 1
// }
//
// func (d *decodeState) init(data []byte) *decodeState {
// 	d.data = data
// 	d.off = 0
// 	d.scanWhile(scanBeginHeader)
// 	if d.lastCode != scanEndHeader {
// 		d.savedError = fmt.Errorf("could not scan header")
// 	}
// 	return d
// }
//
// func (d *decodeState) scanNext() {
// 	if d.off < len(d.data) {
// 		d.lastCode = d.scan.step(&d.scan, d.data[d.off])
// 		d.off++
// 	} else {
// 		d.lastCode = scanEnd
// 		d.off = len(d.data) + 1
// 	}
// }
//
// func (d *decodeState) scanWhile(code int) {
// 	s, data, i := &d.scan, d.data, d.off
// 	for i < len(data) {
// 		newCode := s.step(s, data[i])
// 		i++
// 		if newCode != code {
// 			d.lastCode = newCode
// 			d.off = i
// 			return
// 		}
// 	}
//
// 	d.off = len(data) + 1
// 	d.lastCode = scanEnd
// }
//
// func (d *decodeState) addErrorContext(err error) error {
// 	if d.errorContext != nil && (d.errorContext.Struct != nil || len(d.errorContext.FieldStack) > 0) {
// 		switch err := err.(type) {
// 		case *UnmarshalTypeError:
// 			err.Struct = d.errorContext.Struct.Name()
// 			fieldStack := d.errorContext.FieldStack
// 			if err.Field != "" {
// 				fieldStack = append(fieldStack, err.Field)
// 			}
// 			err.Field = strings.Join(fieldStack, ".")
// 		}
// 	}
// 	return err
// }
//
// func (d *decodeState) saveError(err error) {
// 	if d.savedError == nil {
// 		d.savedError = d.addErrorContext(err)
// 	}
// }
//
// func (d *decodeState) unmarshal(v any) error {
// 	val := reflect.ValueOf(v)
// 	if val.Kind() != reflect.Pointer || val.IsNil() {
// 		return &InvalidUnmarshalError{reflect.TypeOf(v)}
// 	}
// 	d.scan.reset()
// 	// d.scanWhile(someCode)
// 	err := d.value(val)
// 	if err != nil {
// 		return d.addErrorContext(err)
// 	}
// 	return d.savedError
// }
//
// // this is where the fun begins
// func (d *decodeState) value(v reflect.Value) error {
// 	switch d.lastCode {
// 	default:
// 		panic("whooooooooops!!!!!!")
// 	case scanBeginLiteral:
// 		if v.IsValid() {
// 			if err := d.literal(v); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
//
// func (d *decodeState) literal(v reflect.Value) error {
// 	v = indirect(v)
// 	t := v.Type()
//
// 	var fields structFields
//
// 	switch v.Kind() {
// 	case reflect.Map:
// 		return fmt.Errorf("type %v not implemented yet", v.Kind())
// 	case reflect.Struct:
// 		fields = cachedTypeFields(t)
// 	default:
// 		d.saveError(&UnmarshalTypeError{Value: "literal", Type: t, Offset: int64(d.off)})
// 		// maybe we skip to the next delimiter?
// 		return nil
// 	}
// 	var mapElem reflect.Value
// 	var origErrorContext errorContext
// 	if d.errorContext != nil {
// 		origErrorContext = *d.errorContext
// 	}
//
// 	for {
// 		start := d.readIndex()
// 		d.scanWhile(scanBeginLiteral)
// 		end := d.off
// 		item := d.data[start:end]
//
// 		var subv reflect.Value
// 		destring := false
//
// 		if v.Kind() == reflect.Map {
// 			return fmt.Errorf("type %v not implemented yet", v.Kind())
// 		}
// 	}
// 	return nil
// }
//
// type errorContext struct {
// 	Struct     reflect.Type
// 	FieldStack []string
// }
//
// func indirect(v reflect.Value) reflect.Value {
// 	v0 := v
// 	haveAddr := false
//
// 	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
// 		haveAddr = true
// 		v = v.Addr()
// 	}
// 	for {
// 		if v.Kind() == reflect.Interface && !v.IsNil() {
// 			e := v.Elem()
// 			if e.Kind() == reflect.Pointer && !e.IsNil() && e.Elem().Kind() == reflect.Pointer {
// 				haveAddr = false
// 				v = e
// 				continue
// 			}
// 		}
// 		if v.Kind() != reflect.Pointer {
// 			break
// 		}
// 		if v.Elem().Kind() == reflect.Interface && v.Elem().Elem().Equal(v) {
// 			v = v.Elem()
// 			break
// 		}
// 		if v.IsNil() {
// 			v.Set(reflect.New(v.Type().Elem()))
// 		}
// 		if haveAddr {
// 			v = v0
// 			haveAddr = false
// 		} else {
// 			v = v.Elem()
// 		}
// 	}
// 	return v
// }
