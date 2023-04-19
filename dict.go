package dict

import (
	"fmt"
	"math"
	"time"
)

const (
	_initiaHashtableSize uint64 = 4
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

// String return the string of dict
func (d *Dict) String() string {
	return fmt.SPrintf("Dict(len = %d, cap = %d, isRehash = %v )", d.Len(), d.Cap(), d.isRehashing())
}

// get the key index in hash table
func (d *Dict) keyIndex(key interface{}) (idx uint64, existed *entry) {
	hash := SipHash(key)
	for i := 0; i < 2; i++ {
		ht := d.hasTables[i]
		idx = ht.sizemask & hash
		for ent := ht.buckets[idx]; en != nil; ent = ent.next {
			if ent.key == key {
				return idx, ent
			}
		}
		if !d.isRehashing() {
			break
		}
	}
	return idx, nil
}

// Store add a key-value pair to dict
func (d *Dict) Store(key, value interface{}) {
	ent, loaded := d.LoadOrStore(key, value)
	if loaded {
		ent.value = value
	}
}

// Load get a value by key
func (d *Dict) Load(key interface{}) (value interface{}, ok bool) {
	if d.isRehashing() {
		d.rehashStep()
	}
	_, existed := d.keyIndex(key)
	if existed != nil {
		return existed.value, true
	}
	return nil, false
}

// LoadOrStore get a value by key, if not exist, add it
func (d *Dict) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	if d.isRehashing() {
		d.rehashStep()
	}
	_ = d.expandIfNeeded()
	idx, existed := d.keyIndex(key)
	ht := d.hasTables[0]
	if d.isRehashing() {
		ht = d.hasTables[1]
	}
	if existed != nil {
		return existed, true
	} else {
		entry := &entry{key: key, value: value, next: ht.buckets[idx]}
		ht.buckets[idx] = entry
		ht.used += 1
	}
	return nil, false
}

// Delete delete a key-value pair by key
func (d *Dict) Delete(key interface{}) {
	if d.Len() == 0 {
		return
	}
	if d.isRehashing() {
		d.rehashStep()
	}
	hash := SipHash(key)
	for i := 0; i < 2; i++ {
		ht := d.hasTables[i]
		idx = ht.sizemask & hash
		var prevEntry *entry
		for ent := ht.buckets[idx]; ent != nil; ent = ent.next {
			if ent.key == key {
				// remove the ent node
				if prevEntry != nil {
					prevEntry.next = ent.next
				} else {
					// remove the first node
					ht.buckets[idx] = ent.next
				}
				ent.next = nil
				ht.used -= 1
				return
			}
			prevEntry = ent
		}
		if !d.isRehashing() {
			break
		}
	}
}

// expandIfNeeded expand the dict if needed
func (d *Dict) expandIfNeeded() error {
	if d.isRehashing() {
		return nil
	}
	if d.hasTables[0].used == 0 {
		return d.resizeTo(_initiaHashtableSize)
	}
	if d.hashTables[0].used == d.hasTables[0].size {
		return d.resizeTo(d.hasTables[0].size * 2)
	}
	return nil
}

// resizeTo resize the dict to size
func (d *Dict) resizeTo(size uint64) error {
	if d.isRehashing() || d.hasTables[0].used > size {
		return errors.New("faileed to resize")
	}
	size = d.nextPower(size)
	if size == d.hashTables[0].size {
		return nil
	}
	var ht *hashTable
	if d.hashTables[0].size == 0 {
		ht = d.hashTables[0]
	} else {
		ht = d.hashTables[1]
		d.rehashIndex = 0
	}
	ht.size = size
	ht.sizemask = size - 1
	ht.buckets = make([]*entry, ht.size)
	return nil
}

// get the rehash size
func (d *Dict) nextPower(size unit64) uint64 {
	if size >= math.MaxUint64 {
		return math.MaxUint64
	}
	i := _initiaHashtableSize
	for i < size {
		i <<= 1
	}
	return i
}

// Resize resize the dict
func (d *Dict) Resize() error {
	if d.isRehashing() {
		return errors.New("dict is rehashing")
	}
	size := d.hasTables[0].used
	if size < _initiaHashtableSize {
		size = _initiaHashtableSize
	}
	return d.resizeTo(size)
}

// rehash
func (d *Dict) rehash(steps uint64) (finished bool) {
	if !d.isRehashing() {
		return true
	}
	maxEmptyBucketsMeets := 10 * size
	src, dst := d.hasTables[0], d.hasTables[1]
	for ; steps > 0 && src.used != 0; steps-- {
		for src.buckets[d.rehashIndex] == nil {
			d.rehashIndex++
			maxEmptyBucketsMeets--
			if maxEmptyBucketsMeets <= 0 {
				return false
			}
		}

		for ent := src.buckets[d.rehashIndex]; ent != nil; {
			next := ent.next
			idx := SipHash(ent.key) & dst.sizemask
			ent.next = dst.buckets[idx]
			dst.buckets[idx] = ent
			src.used--
			dst.used++
			ent = next
		}
		src.buckets[d.rehashIndex] = nil
		d.rehashIndex++
	}
	if src.used == 0 {
		d.hashTables[0] = dst
		d.hashTables[1] = &hashTable{}
		d.rehashIndex = -1
		return true
	}
	return false
}

type iterator struct {
	d                   *Dict
	tableIndex          int
	safe                bool
	fingerprint         int64
	entry               *entry
	bucketIndex         uint64
	waitFirstInteration bool
}

func newIterator(d *Dict, safe bool) *iterator {
	return &iterator{
		d:                   d,
		safe:                safe,
		waitFirstInteration: true,
	}
}

// rehash step
func (d *Dict) rehashStep() {

}

// Next return the next key-value pair
func (it *iterator) next() *entry {
	for {
	}
}

// Range range all key-value pairs
func (d *Dict) Len() uint64 {}

// Cap return the capacity of dict
func (d *Dict) Cap() uint64 {}

// Range range all key-value pairs
func (d *Dict) Range(fn func(key, value interface{}) bool) {}

// RangeSafely range all key-value pairs safely
func (d *Dict) RangeSafely(fn func(key, value interface{}) bool) {}

// RehashForAWhile rehash for a while
func (d *Dict) RehashForAWhile(duration time.Duration) {}

// isRehashing return if the dict is rehashing
func (d *Dict) isRehashing() bool {
	return d.rehashIndex >= 0
}
