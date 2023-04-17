package dict

import (
	"time"
)

type Dict struct {
	hasTables   []*hashTable
	rehashIndex int64
	iterators   uint64
}

type hashTable struct {
	buckets  []*entry
	size     uint64
	sizemask uint64
	used     uint64
}

type entry struct {
	key, vaule interface{}
	next       *entry
}

// new init a dict
func New() *Dict {
	return &Dict{
		hashTables:  []*hashTable{{}, {}},
		rehashIndex: -1,
		iterators:   0,
	}
}

// Store add a key-value pair to dict
func (d *Dict) Store(key, value interface{}) {}

// Load get a value by key
func (d *Dict) Load(key interface{}) (value interface{}, ok bool) {}

// LoadOrStore get a value by key, if not exist, add it
func (d *Dict) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {}

// Delete delete a key-value pair by key
func (d *Dict) Delete(key interface{}) {}

// Range range all key-value pairs
func (d *Dict) Len() uint64 {}

// Cap return the capacity of dict
func (d *Dict) Cap() uint64 {}

// Range range all key-value pairs
func (d *Dict) Range(fn func(key, value interface{}) bool) {}

// RangeSafely range all key-value pairs safely
func (d *Dict) RangeSafely(fn func(key, value interface{}) bool) {}

// Resize resize the dict
func (d *Dict) Resize() error {}

// RehashForAWhile rehash for a while
func (d *Dict) RehashForAWhile(duration time.Duration) {}
