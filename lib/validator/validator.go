package validator

import (
	"fmt"
	"lib/utils/errs"
	"lib/utils/maps"
	"reflect"
)

var validators = maps.NewWeighted(map[string]maps.Weighted[validatorI]{
	// if `env` weight is above `flags` weight, then
	// even if provided value is empty, but enviromnent variable is set
	// validator will use value from env and required flag will see that value, so it won't complain
	"env": {Value: env{}, Weight: 40},

	"default": {Value: defaultV{}, Weight: 30},

	"flags": {Value: flags{}, Weight: 20},

	"min":   {Value: min{}, Weight: 10},
	"max":   {Value: max{}, Weight: 10},
	"regex": {Value: regex{}, Weight: 10},
})

type validator struct {
	errors errs.Errors
}

func Validate[T any](strct *T) error {
	v := validator{}
	return v.validateStruct("", strct)
}

// strct needs to be pointer to struct
func (v *validator) validateStruct(baseFieldPath string, strct any) error {
	rv := reflect.ValueOf(strct).Elem()
	// needs to resolve interface to call FieldByName
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	// needs to resolve pointer to call FieldByName
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	rt := rv.Type()
	// rt has to be kind struct
	if rt.Kind() != reflect.Struct {
		panic(fmt.Sprintf("validator: cannot validate type %s", rt.Kind()))
	}

	defaultField, ok := rt.FieldByName("_")
	defaultTags := map[string]string{} // tag name -> tag value
	if ok {
		for iter := range validators.Iter() {
			tag := iter.Key
			if val, ok := defaultField.Tag.Lookup(tag); ok {
				defaultTags[tag] = val
			}
		}
	}

	for i := 0; i < rt.NumField(); i++ {
		fieldT := rt.Field(i)
		if fieldT.Name == "_" {
			continue
		}

		fieldV := rv.Field(i)
		resolved := fieldV
		if resolved.Kind() == reflect.Pointer {
			resolved = resolved.Elem()
		}
		if resolved.Kind() == reflect.Struct {
			v.validateStruct(
				fmt.Sprintf("%s%s.", baseFieldPath, fieldT.Name),
				resolved.Addr().Interface(),
			)
		}

		// validate field
		for iter := range validators.Iter() {
			tag := iter.Key
			validator := iter.Value.Value

			if _, ok := defaultTags[tag]; ok {
				// will be validated by default tags, no need to do it now
				continue
			}

			val, ok := fieldT.Tag.Lookup(tag)
			if !ok {
				continue
			}
			if err := validator.validate(val, fieldV); err != nil {
				v.errors = append(v.errors, fmt.Errorf("%s%s: %s", baseFieldPath, fieldT.Name, err.Error()))
			}
		}
		// validate default tags
		for tag, val := range defaultTags {
			validator := validators.Get(tag).Value
			if err := validator.validate(val, fieldV); err != nil {
				v.errors = append(v.errors, fmt.Errorf("%s%s: %s", baseFieldPath, fieldT.Name, err.Error()))
			}
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}
	return nil
}
