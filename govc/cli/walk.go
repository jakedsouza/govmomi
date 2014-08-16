/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"fmt"
	"reflect"
)

type WalkFn func(c interface{}) error

// Walk recursively walks struct types that implement the specified interface.
// Fields that implement the specified interface are expected to be pointer
// values. This allows the function to cache pointer values on a per-type
// basis. If, during a recursive walk, the same type is encountered twice, the
// function creates a new value of that type the first time, and reuses that
// same value the second time.
//
// This function is used to make sure that a hierarchy of flags where multiple
// structs refer to the `Client` flag will not end up with more than one
// instance of the actual client. Rather, every struct referring to the
// `Client` flag will have a pointer to the same underlying `Client` struct.
//
func Walk(c interface{}, ifaceType reflect.Type, fn WalkFn) error {
	var walker WalkFn

	visited := make(map[reflect.Type]reflect.Value)
	walker = func(c interface{}) error {
		v := reflect.ValueOf(c).Elem()
		t := v.Type()

		for i := 0; i < t.NumField(); i++ {
			ff := t.Field(i)
			ft := ff.Type
			fv := v.Field(i)

			// Check that a pointer to this field's type doesn't implement the
			// specified interface. If it does, this field references the type as
			// value. This is not allowed because it prohibits a value from being
			// shared among multiple structs that reference it.
			//
			// For example: if a struct has two fields of the same type, they must
			// both point to the same value after this routine has executed. If these
			// fields are not a pointer type, the value cannot be shared.
			//
			if reflect.PtrTo(ft).Implements(ifaceType) {
				panic(fmt.Sprintf(`field "%s" in struct "%s" must be a pointer`, ff.Name, v.Type()))
			}

			// Type must implement specified interface.
			if !ft.Implements(ifaceType) {
				continue
			}

			// Type must be a pointer.
			if ft.Kind() != reflect.Ptr {
				panic(fmt.Sprintf(`field "%s" in struct "%s" must be a pointer`, ff.Name, v.Type()))
			}

			// Field must not be anonymous.
			//
			// Fields implementing the specified interface may not be anonymous,
			// because it can make its parent struct implement the specified
			// interface implicitly.
			//
			// This, in turn, means that the functions as specified per the interface
			// may be called twice: once on the parent, where the calls as dispatched
			// to this field's type, and once for the field itself.
			//
			if ff.Anonymous {
				panic(fmt.Sprintf(`field "%s" in struct "%s" must not be anonymous`, ff.Name, v.Type()))
			}

			if _, ok := visited[ft]; !ok {
				if fv.IsNil() {
					visited[ft] = reflect.New(ft.Elem())
				} else {
					visited[ft] = fv
				}

				// Not seen before, recurse.
				err := walker(visited[ft].Interface())
				if err != nil {
					return err
				}
			}

			fv.Set(visited[ft])
		}

		// Call user specified function.
		err := fn(c)
		if err != nil {
			return err
		}

		return nil
	}

	return walker(c)
}
