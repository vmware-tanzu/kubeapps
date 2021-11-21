/*
Copyright © 2021 VMware
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
package common

import "reflect"

type Empty struct{}
type T interface{}
type HashSet map[T]Empty

func (s HashSet) Has(item T) bool {
	_, exists := s[item]
	return exists
}

func (s HashSet) Insert(item T) {
	s[item] = Empty{}
}

func (s HashSet) Delete(item T) {
	delete(s, item)
}

func (s HashSet) DeepCopy() HashSet {
	newSet := make(HashSet)
	for k := range s {
		newSet.Insert(k)
	}
	return newSet
}

func (s HashSet) ForEach(f func(T)) {
	for key := range s {
		f(key)
	}
}

func (s HashSet) Values() []reflect.Value {
	return reflect.ValueOf(s).MapKeys()
}

func (s HashSet) IsEmpty() bool {
	return len(s) == 0
}
