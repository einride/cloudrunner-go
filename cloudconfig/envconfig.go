package cloudconfig

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"google.golang.org/grpc/codes"
)

// Copyright (c) 2013 Kelsey Hightower
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Setter is implemented by types can self-deserialize values.
// Any type that implements flag.Value also implements Setter.
type Setter interface {
	Set(value string) error
}

// fieldSpec maintains information about the configuration variable
type fieldSpec struct {
	Name  string
	Key   string
	Value reflect.Value
	Tags  reflect.StructTag
}

func collectFieldSpecs(prefix string, spec interface{}) ([]fieldSpec, error) {
	s := reflect.ValueOf(spec)
	if s.Kind() != reflect.Ptr {
		return nil, errors.New("specification must be a struct pointer")
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, errors.New("specification must be a struct pointer")
	}
	typeOfSpec := s.Type()
	// over allocate an info array, we will extend if needed later
	infos := make([]fieldSpec, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)
		if !f.CanSet() || isTrue(ftype.Tag.Get("ignored")) {
			continue
		}
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					// nil pointer to a non-struct: leave it alone
					break
				}
				// nil pointer to struct: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}
		// Capture information about the config variable
		info := fieldSpec{
			Name:  ftype.Name,
			Value: f,
			Tags:  ftype.Tag,
		}
		// Default to the field name as the env var name (will be upcased)
		info.Key = info.Name
		if prefix != "" {
			info.Key = fmt.Sprintf("%s_%s", prefix, info.Key)
		}
		if ftype.Tag.Get("env") != "" {
			info.Key = strings.ToUpper(ftype.Tag.Get("env"))
		}
		info.Key = strings.ToUpper(info.Key)
		infos = append(infos, info)
		if f.Kind() == reflect.Struct {
			if setterFrom(f) == nil && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil {
				innerPrefix := prefix
				if !ftype.Anonymous {
					innerPrefix = info.Key
				}
				embeddedPtr := f.Addr().Interface()
				embeddedInfos, err := collectFieldSpecs(innerPrefix, embeddedPtr)
				if err != nil {
					return nil, err
				}
				infos = append(infos[:len(infos)-1], embeddedInfos...)
				continue
			}
		}
	}
	return infos, nil
}

func process(fieldSpecs []fieldSpec) error {
	for _, info := range fieldSpecs {
		value, ok := os.LookupEnv(info.Key)
		def := info.Tags.Get("default")
		if !ok && def != "" {
			value = def
		}
		if !ok && metadata.OnGCE() {
			if onGCE := info.Tags.Get("onGCE"); onGCE != "" {
				value = onGCE
				def = onGCE
			}
		}
		if !ok && def == "" {
			if isTrue(info.Tags.Get("required")) {
				key := info.Key
				return fmt.Errorf("required key %s missing value", key)
			}
			continue
		}
		if err := processField(value, info.Value); err != nil {
			return &parseError{
				KeyName:   info.Key,
				FieldName: info.Name,
				TypeName:  info.Value.Type().String(),
				Value:     value,
				Err:       err,
			}
		}
	}
	return nil
}

func processField(value string, field reflect.Value) error {
	typ := field.Type()
	setter := setterFrom(field)
	if setter != nil {
		return setter.Set(value)
	}
	if t := textUnmarshaler(field); t != nil {
		return t.UnmarshalText([]byte(value))
	}
	if b := binaryUnmarshaler(field); b != nil {
		return b.UnmarshalBinary([]byte(value))
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}
	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		if isDuration(field) {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var (
			val uint64
			err error
		)
		if isGRPCCode(field) {
			var c codes.Code
			err = c.UnmarshalJSON([]byte(strconv.Quote(strings.ToUpper(value))))
			val = uint64(c)
		} else {
			val, err = strconv.ParseUint(value, 0, typ.Bits())
		}
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Slice:
		sl := reflect.MakeSlice(typ, 0, 0)
		if typ.Elem().Kind() == reflect.Uint8 {
			sl = reflect.ValueOf([]byte(value))
		} else if len(strings.TrimSpace(value)) != 0 {
			vals := strings.Split(value, ",")
			sl = reflect.MakeSlice(typ, len(vals), len(vals))
			for i, val := range vals {
				err := processField(val, sl.Index(i))
				if err != nil {
					return err
				}
			}
		}
		field.Set(sl)
	case reflect.Map:
		mp := reflect.MakeMap(typ)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, ",")
			for _, pair := range pairs {
				kvpair := strings.Split(pair, ":")
				if len(kvpair) != 2 {
					return fmt.Errorf("invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				err := processField(kvpair[0], k)
				if err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				err = processField(kvpair[1], v)
				if err != nil {
					return err
				}
				mp.SetMapIndex(k, v)
			}
		}
		field.Set(mp)
	}
	return nil
}

func interfaceFrom(field reflect.Value, fn func(interface{}, *bool)) {
	// it may be impossible for a struct field to fail this check
	if !field.CanInterface() {
		return
	}
	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}

func setterFrom(field reflect.Value) (s Setter) {
	interfaceFrom(field, func(v interface{}, ok *bool) { s, *ok = v.(Setter) })
	return s
}

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { t, *ok = v.(encoding.TextUnmarshaler) })
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { b, *ok = v.(encoding.BinaryUnmarshaler) })
	return b
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func isGRPCCode(value reflect.Value) bool {
	return value.Kind() == reflect.Uint32 &&
		value.Type().PkgPath() == "google.golang.org/grpc/codes" &&
		value.Type().Name() == "Code"
}

func isDuration(value reflect.Value) bool {
	return value.Kind() == reflect.Int64 &&
		value.Type().PkgPath() == "time" &&
		value.Type().Name() == "Duration"
}

// A parseError occurs when an environment variable cannot be converted to
// the type required by a struct field during assignment.
type parseError struct {
	KeyName   string
	FieldName string
	TypeName  string
	Value     string
	Err       error
}

func (e *parseError) Error() string {
	return fmt.Sprintf(
		"config parse error: assigning %[1]s to %[2]s: converting '%[3]s' to type %[4]s. details: %[5]s",
		e.KeyName,
		e.FieldName,
		e.Value,
		e.TypeName,
		e.Err,
	)
}
