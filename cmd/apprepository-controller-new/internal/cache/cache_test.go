/*
Copyright 2022 The Flux authors

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

package cache

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestCache(t *testing.T) {
	g := NewWithT(t)
	// create a cache that can hold 2 items and have no cleanup
	cache := New(2, 0)

	// Get an Item from the cache
	if _, found := cache.Get("key1"); found {
		t.Error("Item should not be found")
	}

	// Add an item to the cache
	err := cache.Add("key1", "value1", 0)
	g.Expect(err).ToNot(HaveOccurred())

	// Get the item from the cache
	item, found := cache.Get("key1")
	g.Expect(found).To(BeTrue())
	g.Expect(item).To(Equal("value1"))

	// Add another item to the cache
	err = cache.Add("key2", "value2", 0)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cache.ItemCount()).To(Equal(2))

	// Get the item from the cache
	item, found = cache.Get("key2")
	g.Expect(found).To(BeTrue())
	g.Expect(item).To(Equal("value2"))

	//Add an item to the cache
	err = cache.Add("key3", "value3", 0)
	g.Expect(err).To(HaveOccurred())

	// Replace an item in the cache
	err = cache.Set("key2", "value3", 0)
	g.Expect(err).ToNot(HaveOccurred())

	// Get the item from the cache
	item, found = cache.Get("key2")
	g.Expect(found).To(BeTrue())
	g.Expect(item).To(Equal("value3"))

	// new cache with a cleanup interval of 1 second
	cache = New(2, 1*time.Second)

	// Add an item to the cache
	err = cache.Add("key1", "value1", 2*time.Second)
	g.Expect(err).ToNot(HaveOccurred())

	// Get the item from the cache
	item, found = cache.Get("key1")
	g.Expect(found).To(BeTrue())
	g.Expect(item).To(Equal("value1"))

	// wait for the item to expire
	time.Sleep(3 * time.Second)

	// Get the item from the cache
	item, found = cache.Get("key1")
	g.Expect(found).To(BeFalse())
	g.Expect(item).To(BeNil())
}
