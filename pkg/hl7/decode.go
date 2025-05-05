package hl7

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var DefaultSegDelim = byte('\r')

func Unmarshal(data []byte, v any) error {
	d := newDecoder()
	d.init(data, DefaultSegDelim)
	if d.savedError != nil {
		return d.savedError
	}
	return d.unmarshal(v)
	// return nil
}

func NewDecoder(data []byte) *Decoder {
	d := newDecoder()
	d.init(data, DefaultSegDelim)
	return d
}

func (d *Decoder) Decode(v any) error {
	return d.unmarshal(v)
}

type Decoder struct {
	segMap       segmentMap       // this converts the name to the (list of) zero-based idx
	segments     map[int]*segment // key is zero-based idx of segment
	data         []byte
	start, off   int
	lastState    scanState
	scan         *scanner
	errorContext *errorContext
	savedError   error
}

func newDecoder() *Decoder {
	return &Decoder{
		segMap:   newSegmentMap(),
		segments: make(map[int]*segment),
	}
}

func (d *Decoder) init(data []byte, segDelim byte) {
	if len(data) < 8 {
		d.saveError(fmt.Errorf("message is too short (length: %d)\n", len(data)))
		return
	}
	d.data = data
	d.lastState = stateBegin
	d.scan = &scanner{
		step:     scanSegmentName,
		segDelim: segDelim,
		fldDelim: data[3],
		comDelim: data[4],
		repDelim: data[5],
		escDelim: data[6],
		subDelim: data[7],
	} // switch to pool

	var (
		idx1             = 1
		currentSegIdx    int
		currentSegName   string
		currentSegFields *segment
	)

	for d.off < len(d.data) {
		switch d.lastState {
		case stateBegin:
			d.scanWhile(stateSegmentName)
			if d.lastState != stateSegmentNameEnd {
				panic("huh!?")
			}
			currentSegIdx = d.start
			currentSegName = string(d.data[currentSegIdx:d.readIndex()])
			currentSegFields = NewSegment(currentSegName)
			if currentSegName != messageHeader {
				d.saveError(fmt.Errorf("first segment should be 'MSH', got %s", currentSegName))
				return
			}
			idx1++
		case stateSegmentNameEnd:
			if ok := currentSegFields.AddField(idx1, d.off, d.scan); ok {
				panic("didn't expect to find duplicate field")
			}
			idx1++
			d.scanNext()
		case stateSegmentName:
			d.scanWhile(stateSegmentName)
			if d.lastState != stateSegmentNameEnd {
				panic("expected end of reading segment name")
			}
			currentSegIdx = d.start
			currentSegName = string(d.data[currentSegIdx:d.readIndex()])
			currentSegFields = NewSegment(currentSegName)
		case stateSegmentEnd:
			currentSegFields.endIdx = d.readIndex()
			d.segMap.addSegment(currentSegName, currentSegIdx)
			d.segments[currentSegIdx] = currentSegFields
			d.start = d.off
			idx1 = 1
			d.scanNext()
		case stateFieldValEnd:
			if ok := currentSegFields.AddField(idx1, d.off, d.scan); ok {
				panic("didn't expect to find duplicate field")
			}
			idx1++
			d.scanNext()
		case stateContinue:
			d.scanWhile(stateContinue)
		default:
			panic(fmt.Sprintf("unrecognized state! got: '%d'\n", d.lastState))
		}
	}
	currentSegFields.AddField(idx1, d.off+1, d.scan)
	d.segMap.addSegment(currentSegName, d.start)
	d.segments[currentSegIdx] = currentSegFields
}

func (d *Decoder) scanNext() {
	if d.off < len(d.data) {
		d.lastState = d.scan.step(d.scan, d.data[d.off])
		d.off++
	}
}

func (d *Decoder) scanWhile(state scanState) {
	s, data, i := d.scan, d.data, d.off
	for i < len(data) {
		newState := s.step(s, d.data[i])
		i++
		if newState != state {
			d.lastState = newState
			d.off = i
			return
		}
	}
	d.off = len(d.data)
	d.lastState = stateErr
}

// n is the "nth" segment repeat
func (d *Decoder) getFieldVal(s string, idx1, n int) string {
	if s == messageHeader {
		switch idx1 {
		case 1:
			return string(d.scan.fldDelim)
		case 2:
			return fmt.Sprintf("%c%c%c%c", d.scan.comDelim, d.scan.repDelim, d.scan.escDelim, d.scan.subDelim)
		}
	}
	indices, found := d.segMap.getSegmentIndices(s)
	if !found || n >= len(indices) {
		return ""
	}
	return d.scanField(idx1, indices[n])
}

func (d *Decoder) scanField(idx1, idx0 int) string {
	segment, ok := d.segments[idx0]
	if !ok {
		return ""
	}
	field, exists := segment.fields.getFieldNode(idx1)
	if !exists {
		return ""
	}
	start := field.idx
	end := segment.endIdx
	if field.next != nil {
		end = field.next.idx - 1
	}
	return string(d.data[start:end])
}

type UnmarshalTypeError struct {
	Value  string
	Type   reflect.Type
	Offset int64
	Struct string
	Field  string
}

func (e *UnmarshalTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "hl7: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "hl7: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "hl7: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "hl7: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "hl7: Unmarshal(nil " + e.Type.String() + ")"
}

type field struct {
	name      string
	nameBytes []byte
	tag       bool
	idx       []int
	typ       reflect.Type
}

type structFields struct {
	list   []field
	byName map[string]*field
}

var fieldCache sync.Map

func typeFields(t reflect.Type) structFields {
	current := []field{}
	next := []field{{typ: t}}

	var count, nextCount map[reflect.Type]int
	visited := map[reflect.Type]bool{}

	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			for i := range f.typ.NumField() {
				sf := f.typ.Field(i)
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Pointer {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						continue
					}
				} else if !sf.IsExported() {
					continue
				}
				tag := sf.Tag.Get("hl7")
				if tag == "-" {
					continue
				}
				name, _ := parseTag(tag) // may change this to return variable list of tags (e.g. ORC,OBR)
				if !isValidTag(name) {
					name = ""
				}
				idx := make([]int, len(f.idx)+1)
				copy(idx, f.idx)
				idx[len(f.idx)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}

				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name: name,
						tag:  tagged,
						idx:  idx,
						typ:  ft,
					}
					field.nameBytes = []byte(field.name)
					fields = append(fields, field)
					if count[f.typ] > 1 {
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), idx: idx, typ: ft})
				}
			}
		}
	}
	slices.SortFunc(fields, func(a, b field) int {
		if c := strings.Compare(a.name, b.name); c != 0 {
			return c
		}
		if c := cmp.Compare(len(a.idx), len(b.idx)); c != 0 {
			return c
		}
		if a.tag != b.tag {
			if a.tag {
				return -1
			}
			return +1
		}
		return slices.Compare(a.idx, b.idx)
	})
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 {
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}
	fields = out
	slices.SortFunc(fields, func(i, j field) int {
		return slices.Compare(i.idx, j.idx)
	})

	exactNameIdx := make(map[string]*field, len(fields))
	for i, field := range fields {
		exactNameIdx[field.name] = &fields[i]
	}
	return structFields{fields, exactNameIdx}
}

func dominantField(fields []field) (field, bool) {
	if len(fields) > 1 && len(fields[0].idx) == len(fields[1].idx) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}

func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structFields)
}

// type decodeState struct {
// 	data         []byte
// 	off          int
// 	lastCode     int
// 	scan         scanner
// 	errorContext *errorContext
// 	savedError   error
// }

func (d *Decoder) readIndex() int {
	return d.off - 1
}

func (d *Decoder) addErrorContext(err error) error {
	if d.errorContext != nil && (d.errorContext.Struct != nil || len(d.errorContext.FieldStack) > 0) {
		switch err := err.(type) {
		case *UnmarshalTypeError:
			err.Struct = d.errorContext.Struct.Name()
			fieldStack := d.errorContext.FieldStack
			if err.Field != "" {
				fieldStack = append(fieldStack, err.Field)
			}
			err.Field = strings.Join(fieldStack, ".")
		}
	}
	return err
}

func (d *Decoder) saveError(err error) {
	if d.savedError == nil {
		d.savedError = d.addErrorContext(err)
	}
}

func (d *Decoder) unmarshal(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	// d.scanWhile(someCode)
	err := d.value(val, 0)
	if err != nil {
		return d.addErrorContext(err)
	}
	return d.savedError
}

func (d *Decoder) repeatSegments(v reflect.Value, elemType reflect.Type) error {
	for i := 0; ; i++ {
		elemPtr := reflect.New(elemType).Elem()
		err := d.prep(elemPtr, elemType, i)
		if err != nil {
			return err
		}
		if isEmptyStruct(elemPtr) {
			break
		}
		v.Set(reflect.Append(v, elemPtr))
	}
	return nil
}

func isEmptyStruct(v reflect.Value) bool {
	for i := range v.NumField() {
		field := v.Field(i)
		if field.IsValid() && !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			return false
		}
	}
	return true
}

func (d *Decoder) value(v reflect.Value, idx int) error {
	val := v.Elem()
	typ := val.Type()

	if val.Kind() == reflect.Slice {
		elemType := typ.Elem()
		if elemType.Kind() != reflect.Struct {
			return fmt.Errorf("expected a slice of structs, but got %s", elemType)
		}
		return d.repeatSegments(val, elemType)
	}

	return d.prep(val, typ, idx)
}

func (d *Decoder) prep(val reflect.Value, typ reflect.Type, idx int) error {
	for i := range typ.NumField() {
		field := typ.Field(i)
		hl7Tag := field.Tag.Get("hl7")
		if hl7Tag == "" {
			continue
		}
		parts := strings.Split(hl7Tag, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid tag: %s", hl7Tag)
		}
		fNum, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		strVal := d.getFieldVal(parts[0], fNum, idx)
		setField(val.Field(i), field.Type, strVal)
	}
	return nil
}

func setField(fVal reflect.Value, fType reflect.Type, val string) {
	if val == "" {
		return
	}
	if fVal.Kind() == reflect.Ptr && fVal.IsNil() {
		panic("value must be a nil pointer!")
	}
	switch fVal.Kind() {
	case reflect.String:
		fVal.SetString(val)
	case reflect.Struct:
		components := strings.Split(val, "^")
		if len(components) == 1 {
			field := fVal.FieldByIndex([]int{0})
			field.SetString(val)
			return
		}
		for i := range fType.NumField() {
			structField := fType.Field(i)
			hl7Tag := structField.Tag.Get("hl7")
			if hl7Tag == "" {
				continue
			}
			idx, err := strconv.Atoi(hl7Tag)
			if err != nil || idx > len(components) {
				continue
			}
			fieldVal := fVal.Field(i)
			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(components[idx-1])
			case reflect.Struct:
				subComps := strings.Split(components[idx-1], "&")
				if len(components) == 1 {
					subField := fieldVal.FieldByIndex([]int{0})
					subField.SetString(components[idx-1])
					continue
				}
				subTyp := fieldVal.Type()
				for j := range subTyp.NumField() {
					subField := fieldVal.Type().Field(j)
					subHl7Tag := subField.Tag.Get("hl7")
					if subHl7Tag == "" {
						continue
					}
					subIdx, err := strconv.Atoi(subHl7Tag)
					if err != nil || subIdx > len(subComps) {
						continue
					}
					subVal := fieldVal.Field(j)
					subVal.SetString(subComps[subIdx-1])
				}
			}
		}
	case reflect.Slice:
		if fVal.IsNil() {
			fVal.Set(reflect.MakeSlice(fVal.Type(), 0, 0))
		}
		repeats := strings.Split(val, "~")
		for _, repeat := range repeats {
			if fVal.Type().Elem().Kind() == reflect.Struct {
				elem := reflect.New(fVal.Type().Elem()).Elem()
				setField(elem, fVal.Type().Elem(), repeat)
				fVal.Set(reflect.Append(fVal, elem))
			} else {
				fVal.Set(reflect.Append(fVal, reflect.ValueOf(repeat)))
			}
		}
	default:
		panic("unsupported type!")
	}
}

type errorContext struct {
	Struct     reflect.Type
	FieldStack []string
}

func indirect(v reflect.Value) reflect.Value {
	v0 := v
	haveAddr := false

	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Pointer && !e.IsNil() && e.Elem().Kind() == reflect.Pointer {
				haveAddr = false
				v = e
				continue
			}
		}
		if v.Kind() != reflect.Pointer {
			break
		}
		if v.Elem().Kind() == reflect.Interface && v.Elem().Elem().Equal(v) {
			v = v.Elem()
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if haveAddr {
			v = v0
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return v
}
