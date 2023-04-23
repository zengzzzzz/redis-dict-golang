/*
 * @Author: zengzh 
 * @Date: 2023-04-23 19:48:01 
 * @Last Modified by:   zengzh 
 * @Last Modified time: 2023-04-23 19:48:01 
 */
package dict

import (
	"errors"
	"fmt"
	"math"
	"time"
	"unsafe"
)

const (
	_initiaHashtableSize uint64 = 4
)

type Dict struct {
	hashTables  []*hashTable
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
	key, value interface{}
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
	return fmt.Sprintf("Dict(len = %d, cap = %d, isRehash = %v )", d.Len(), d.Cap(), d.isRehashing())
}

// get the key index in hash table
func (d *Dict) keyIndex(key interface{}) (idx uint64, existed *entry) {
	hash := SipHash(key)
	for i := 0; i < 2; i++ {
		ht := d.hashTables[i]
		idx = ht.sizemask & hash
		if idx > ht.size {
			return idx, nil
		}
		for ent := ht.buckets[idx]; ent != nil; ent = ent.next {
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
	ent, loaded := d.loadOrStore(key, value)
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
func (d *Dict) loadOrStore(key, value interface{}) (ent *entry, loaded bool) {
	if d.isRehashing() {
		d.rehashStep()
	}
	_ = d.expandIfNeeded()
	idx, existed := d.keyIndex(key)
	ht := d.hashTables[0]
	if d.isRehashing() {
		ht = d.hashTables[1]
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
		ht := d.hashTables[i]
		idx := ht.sizemask & hash
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
	if d.hashTables[0].used == 0 {
		return d.resizeTo(_initiaHashtableSize)
	}
	if d.hashTables[0].used == d.hashTables[0].size {
		return d.resizeTo(d.hashTables[0].size * 2)
	}
	return nil
}

// resizeTo resize the dict to size
func (d *Dict) resizeTo(size uint64) error {
	if d.isRehashing() || d.hashTables[0].used > size {
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
func (d *Dict) nextPower(size uint64) uint64 {
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
	size := d.hashTables[0].used
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
	maxEmptyBucketsMeets := 10 * steps
	src, dst := d.hashTables[0], d.hashTables[1]
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
	if d.iterators == 0 {
		d.rehash(1)
	}
}

// Next return the next key-value pair
func (it *iterator) next() *entry {
	for {
		if it.entry == nil {
			if it.waitFirstInteration {
				if it.safe {
					it.d.iterators++
				} else {
					it.fingerprint = it.d.fingerprint()
				}
			}
			ht := it.d.hashTables[it.tableIndex]
			if it.bucketIndex >= ht.size {
				if !it.d.isRehashing() || it.tableIndex != 0 {
					return nil
				}
				it.tableIndex = 1
				it.bucketIndex = 0
				ht = it.d.hashTables[1]
			}
			it.entry = ht.buckets[it.bucketIndex]
			it.bucketIndex++
		} else {
			it.entry = it.entry.next
		}
		if it.entry != nil {
			return it.entry
		}
	}
}

// Release release the iterator
func (it *iterator) release() {
	if it.safe {
		it.d.iterators--
	} else {
		fp := it.d.fingerprint()
		if fp != it.fingerprint {
			panic("operations like 'LoadOrStore', 'Load' or 'Delete' are not safe for an unsafe iterator")
		}
	}
}

// Range range all key-value pairs
func (d *Dict) rangeDict(fn func(key, value interface{}) bool, safe bool) {
	it := newIterator(d, safe)
	defer it.release()
	for {
		if ent := it.next(); ent != nil {
			if !fn(ent.key, ent.value) {
				break
			}
		} else {
			break
		}
	}
}

// fingerprint return the fingerprint of the dict
func (d *Dict) fingerprint() int64 {
	metas := []int64{
		// meta of table 0
		int64(uintptr(unsafe.Pointer(&d.hashTables[0].buckets))),
		int64(d.hashTables[0].size),
		int64(d.hashTables[0].used),
		// meta of table 1
		int64(uintptr(unsafe.Pointer(&d.hashTables[1].buckets))),
		int64(d.hashTables[1].size),
		int64(d.hashTables[1].used),
	}

	var hash int64
	for _, meta := range metas {
		hash += meta
		// 使用 Tomas Wang 64 位整数 hash 算法
		hash = (hash << 21) - hash - 1
		hash = hash ^ (hash >> 24)
		hash = (hash + (hash << 3)) + (hash << 8) // hash * 256
		hash = hash ^ (hash >> 14)
		hash = (hash + (hash << 2)) + (hash << 4) // hash * 21
		hash = hash ^ (hash >> 28)
		hash = hash + (hash << 31)
	}

	return hash
}

// Range range all key-value pairs
func (d *Dict) Len() uint64 {
	var _len uint64
	for _, ht := range d.hashTables {
		_len += ht.used
	}
	return _len
}

// Cap return the capacity of dict
func (d *Dict) Cap() uint64 {
	if d.isRehashing() {
		return d.hashTables[1].size
	}
	return d.hashTables[0].size
}

// Range range all key-value pairs
func (d *Dict) Range(fn func(key, value interface{}) bool) {
	d.rangeDict(fn, false)
}

// RangeSafely range all key-value pairs safely
func (d *Dict) RangeSafely(fn func(key, value interface{}) bool) {
	d.rangeDict(fn, true)
}

// RehashForAWhile rehash for a while
func (d *Dict) RehashForAWhile(duration time.Duration) int64 {
	tm := time.NewTimer(duration)
	defer tm.Stop()
	var rehashes int64
	for {
		select {
		case <-tm.C:
			return rehashes
		default:
			if d.rehash(100) {
				return rehashes
			}
			rehashes += 100
		}
	}
}

// isRehashing return if the dict is rehashing
func (d *Dict) isRehashing() bool {
	return d.rehashIndex >= 0
}
