package gitc

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// https://git-scm.com/docs/pretty-formats

// Output need to be formated as: []byte(fmt.Sprintf("[%s]", string(out[:len(out)-1])))
func PrettyJson[T any]() string {
	const fieldTmpl = `"%s": %s`
	obj := (*T)(nil)
	typ := reflect.TypeOf(obj).Elem()

	lines := make([]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("git")
		if tag == "" {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Int64, reflect.TypeOf((*time.Time)(nil)).Elem().Kind():
			lines[i] = fmt.Sprintf(fieldTmpl, field.Name, tag)
		default:
			lines[i] = fmt.Sprintf(fieldTmpl, field.Name, `"`+tag+`"`)
		}
	}
	return fmt.Sprintf("{%s},", strings.Join(lines, ","))
}
