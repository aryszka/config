package config

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/aryszka/config/keys"
)

func tooManyValues(...interface{}) error {
	return ErrTooManyValues
}

func invalidBooleanValue(...interface{}) error {
	return ErrInvalidInputValue
}

func invalidNumericValue(a ...interface{}) error {
	if len(a) > 0 {
		if err, _ := a[0].(error); errors.Is(err, ErrInvalidInputValue) {
			return err
		}
	}

	return ErrInvalidInputValue
}

func overflow(...interface{}) error {
	return ErrNumericOverflow
}

func invalidStringValue(...interface{}) error {
	return ErrInvalidInputValue
}

func invalidStructureValue(...interface{}) error {
	return ErrInvalidInputValue
}

func invalidListValue(...interface{}) error {
	return ErrInvalidInputValue
}

func invalidMapType(...interface{}) error {
	return ErrInvalidTarget
}

func multipleCanonicalKeys(...interface{}) error {
	return ErrConflictingKeys
}

func invalidType(...interface{}) error {
	return ErrInvalidInputValue
}

func invalidTarget(...interface{}) error {
	return ErrInvalidTarget
}

func zeroOrOne(apply func(reflect.Value, Node) (bool, error), v reflect.Value, n Node) (bool, error) {
	if n.Len() > 1 {
		return false, tooManyValues()
	}

	if n.Len() == 0 {
		return false, nil
	}

	return apply(v, n.Item(0))
}

func parseUint(s string) (uint64, error) {
	var base int
	switch {
	case strings.HasPrefix(s, "0x"):
		if len(s) == 2 {
			return 0, invalidNumericValue(s)
		}

		base = 16
		s = s[2:]
	case strings.HasPrefix(s, "0"):
		if s == "0" {
			return 0, nil
		}

		base = 8
		s = s[1:]
	default:
		base = 10
	}

	return strconv.ParseUint(s, base, 64)
}

func parseInt(s string) (int64, error) {
	var negative bool
	if strings.HasPrefix(s, "-") {
		if len(s) == 1 {
			return 0, invalidNumericValue(s)
		}

		negative = true
		s = s[1:]
	}

	u, err := parseUint(s)
	if err != nil {
		return 0, err
	}

	i := int64(u)
	if i < 0 && (!negative || i > 0-int64(^uint64(0)>>1)-1) {
		return 0, overflow(s)
	}

	if negative {
		// itself if min negative
		i = 0 - i
	}

	return i, nil
}

func applyBool(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&Bool == 0 {
		return false, invalidBooleanValue()
	}

	if t&List != 0 {
		return zeroOrOne(applyBool, v, n)
	}

	switch tvalue := n.Primitive().(type) {
	case bool:
		v.SetBool(tvalue)
		return true, nil
	case string:
		b, err := strconv.ParseBool(tvalue)
		if err != nil {
			return false, invalidBooleanValue(err)
		}

		v.SetBool(b)
		return true, nil
	default:
		return false, ErrSourceImplementation
	}
}

func applyInt(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&Int == 0 {
		return false, invalidNumericValue()
	}

	if t&List != 0 {
		return zeroOrOne(applyInt, v, n)
	}

	value := n.Primitive()
	if svalue, ok := value.(string); ok {
		i, err := parseInt(svalue)
		if err != nil {
			return false, invalidNumericValue(err)
		}

		if v.OverflowInt(i) {
			return false, overflow(value)
		}

		v.SetInt(i)
		return true, nil
	}

	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		ri := rvalue.Int()
		if v.OverflowInt(ri) {
			return false, overflow(value)
		}

		v.SetInt(ri)
		return true, nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		ru := rvalue.Uint()
		ri := int64(ru)
		if ri < 0 || v.OverflowInt(ri) {
			return false, overflow(value)
		}

		v.SetInt(ri)
		return true, nil
	case reflect.Float32, reflect.Float64:
		rf := rvalue.Float()
		ri := int64(rf)
		if float64(ri) != rf {
			return false, invalidNumericValue(value)
		}

		if v.OverflowInt(ri) {
			return false, overflow(value)
		}

		v.SetInt(ri)
		return true, nil
	default:
		return false, ErrSourceImplementation
	}
}

func applyUint(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&Int == 0 {
		return false, invalidNumericValue()
	}

	if t&List != 0 {
		return zeroOrOne(applyUint, v, n)
	}

	value := n.Primitive()
	if svalue, ok := value.(string); ok {
		u, err := parseUint(svalue)
		if err != nil {
			return false, invalidNumericValue(err)
		}

		if v.OverflowUint(u) {
			return false, overflow(value)
		}

		v.SetUint(u)
		return true, nil
	}

	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		ri := rvalue.Int()
		if ri < 0 {
			return false, overflow(value)
		}

		ru := uint64(ri)
		if v.OverflowUint(ru) {
			return false, overflow(value)
		}

		v.SetUint(ru)
		return true, nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		ru := rvalue.Uint()
		if v.OverflowUint(ru) {
			return false, overflow(value)
		}

		v.SetUint(ru)
		return true, nil
	case reflect.Float32, reflect.Float64:
		rf := rvalue.Float()
		if rf < 0 {
			return false, overflow(value)
		}

		ru := uint64(rf)
		if float64(ru) != rf {
			return false, invalidNumericValue(value)
		}

		if v.OverflowUint(ru) {
			return false, overflow(value)
		}

		v.SetUint(ru)
		return true, nil
	default:
		return false, ErrSourceImplementation
	}
}

func applyFloat(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&Float == 0 {
		return false, invalidNumericValue()
	}

	value := n.Primitive()
	if svalue, ok := value.(string); ok {
		f, err := strconv.ParseFloat(svalue, v.Type().Bits())
		if err != nil {
			return false, invalidNumericValue(err)
		}

		v.SetFloat(f)
		return true, nil
	}

	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		ri := rvalue.Int()
		rf := float64(ri)
		v.SetFloat(rf)
		return true, nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		ru := rvalue.Uint()
		rf := float64(ru)
		v.SetFloat(rf)
		return true, nil
	case reflect.Float32, reflect.Float64:
		rf := rvalue.Float()
		if v.OverflowFloat(rf) {
			return false, overflow(value)
		}

		v.SetFloat(rf)
		return true, nil
	default:
		return false, ErrSourceImplementation
	}
}

func applyString(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&String == 0 {
		return false, invalidStringValue()
	}

	if t&List != 0 {
		return zeroOrOne(applyString, v, n)
	}

	value := n.Primitive()
	svalue, ok := value.(string)
	if !ok {
		return false, invalidStringValue(value)
	}

	v.SetString(svalue)
	return true, nil
}

func exported(name string) bool {
	return unicode.IsUpper([]rune(name)[0])
}

func applyStruct(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t&Structure == 0 {
		return false, invalidStructureValue()
	}

	canonicalKeys := make(map[string]string)
	for _, key := range n.Keys() {
		canonical := keys.CanonicalSymbol(key)
		if _, has := canonicalKeys[canonical]; has {
			// TODO: this decision should be made in the source or the reader
			return false, multipleCanonicalKeys(canonical)
		}

		canonicalKeys[canonical] = key
	}

	var set bool
	vt := v.Type()
	for i := 0; i < vt.NumField(); i++ {
		f := vt.Field(i)
		if !exported(f.Name) {
			continue
		}

		key, ok := canonicalKeys[keys.CanonicalSymbol(f.Name)]
		if !ok {
			continue
		}

		if isSet, err := apply(v.FieldByName(f.Name), n.Field(key)); err != nil {
			return set, err
		} else if isSet {
			set = true
		}
	}

	return set, nil
}

func applyMap(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if v.Type().Key().Kind() != reflect.String {
		return false, invalidMapType()
	}

	if t == Nil {
		v.Set(reflect.Zero(v.Type()))
		return true, nil
	}

	if t&Structure == 0 {
		return false, invalidStructureValue()
	}

	keys := n.Keys()
	if len(keys) == 0 {
		return false, nil
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	for _, key := range keys {
		pfv := reflect.New(v.Type().Elem())
		if _, err := apply(pfv, n.Field(key)); err != nil {
			return true, err
		}

		v.SetMapIndex(reflect.ValueOf(key), pfv.Elem())
	}

	return true, nil
}

func applyList(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t == Nil {
		v.Set(reflect.Zero(v.Type()))
		return true, nil
	}

	if t&List == 0 {
		return false, invalidListValue()
	}

	l := n.Len()
	v.Set(reflect.MakeSlice(v.Type(), l, l))
	if n.Len() == 0 {
		return false, nil
	}

	for i := 0; i < l; i++ {
		piv := reflect.New(v.Type().Elem())
		if _, err := apply(piv, n.Item(i)); err != nil {
			return true, err
		}

		v.Index(i).Set(piv.Elem())
	}

	return true, nil
}

func applyInterface(v reflect.Value, n Node) (bool, error) {
	t := n.Type()

	if t == Nil {
		v.Set(reflect.Zero(v.Type()))
		return true, nil
	}

	switch {
	case t&Primitive != 0 && (t&List == 0 || n.Len() <= 1):
		t := reflect.TypeOf(n.Primitive())
		if !t.Implements(v.Type()) {
			return false, invalidType()
		}

		v.Set(reflect.ValueOf(n.Primitive()))
		return true, nil
	case t&List != 0:
		t := reflect.TypeOf([]interface{}{})
		if !t.Implements(v.Type()) {
			return false, invalidType()
		}

		pv := reflect.New(t)
		set, err := apply(pv, n)
		if set {
			v.Set(pv.Elem())
		}

		return set, err
	case t&Structure != 0:
		m := map[string]interface{}{}
		t := reflect.TypeOf(m)
		if !t.Implements(v.Type()) {
			return false, invalidType()
		}

		vv := reflect.ValueOf(m)
		set, err := apply(vv, n)
		if set {
			v.Set(vv)
		}

		return set, err
	default:
		return false, nil
	}
}

func applyPointer(v reflect.Value, n Node) (bool, error) {
	pv := v
	if pv.IsNil() {
		pv = reflect.New(v.Type().Elem())
	}

	set, err := apply(pv.Elem(), n)
	if !set || err != nil {
		return false, err
	}

	if v.IsNil() {
		v.Set(pv)
	}

	return true, nil
}

func apply(v reflect.Value, n Node) (bool, error) {
	// TODO: check here if implements config parser

	switch v.Kind() {
	case reflect.Bool:
		return applyBool(v, n)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return applyInt(v, n)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return applyUint(v, n)
	case reflect.Float32, reflect.Float64:
		return applyFloat(v, n)
	case reflect.String:
		return applyString(v, n)
	case reflect.Struct:
		return applyStruct(v, n)
	case reflect.Map:
		return applyMap(v, n)
	case reflect.Slice:
		return applyList(v, n)
	case reflect.Interface:
		return applyInterface(v, n)
	case reflect.Ptr:
		return applyPointer(v, n)
	default:
		return false, invalidTarget()
	}
}

// It may change the target even if fails.
func Apply(applyTo interface{}, s Source) error {
	v := reflect.ValueOf(applyTo)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return invalidTarget()
	}

	n, err := s.Read()
	if errors.Is(err, ErrNoConfig) {
		return nil
	}

	if err != nil {
		return err
	}

	_, err = apply(v, n)
	return err
}
