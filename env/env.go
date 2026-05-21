// Package env loads configuration into struct fields from environment
// variables. The struct's fields are annotated with `env:"NAME"` tags;
// supported options are: NAME, NAME,default=VAL, NAME,required.
//
// Supported field types: string, bool, int/intN, uint/uintN, float32,
// float64, time.Duration, []string (comma-separated). Pointers to these
// are also accepted: a missing env var leaves the pointer nil.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Load populates the fields of *dst from os.Getenv based on `env` tags.
// Returns an aggregated error if any required vars are missing or any
// value cannot be parsed.
func Load(dst any) error {
	return load(dst, os.Getenv)
}

// LoadFromMap is Load that draws from a provided map (useful in tests).
func LoadFromMap(dst any, m map[string]string) error {
	return load(dst, func(k string) string { return m[k] })
}

func load(dst any, get func(string) string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("env: dst must be *struct, got %T", dst)
	}
	v = v.Elem()
	t := v.Type()

	var errs []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("env")
		if !ok {
			continue
		}
		name, def, required := parseTag(tag)

		raw := get(name)
		if raw == "" {
			if required {
				errs = append(errs, fmt.Sprintf("missing required env %q", name))
				continue
			}
			raw = def
		}
		if raw == "" {
			continue
		}
		if err := setField(v.Field(i), raw); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("env: %s", strings.Join(errs, "; "))
	}
	return nil
}

func parseTag(tag string) (name, def string, required bool) {
	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		switch {
		case p == "required":
			required = true
		case strings.HasPrefix(p, "default="):
			def = strings.TrimPrefix(p, "default=")
		}
	}
	return
}

func setField(fv reflect.Value, raw string) error {
	if fv.Kind() == reflect.Pointer {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return setField(fv.Elem(), raw)
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fv.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return err
			}
			fv.SetInt(int64(d))
			return nil
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		fv.SetFloat(f)
	case reflect.Slice:
		if fv.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("only []string is supported")
		}
		parts := strings.Split(raw, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		fv.Set(reflect.ValueOf(out))
	default:
		return fmt.Errorf("unsupported kind %s", fv.Kind())
	}
	return nil
}
