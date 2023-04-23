/*
 * @Author: zengzh
 * @Date: 2023-04-23 14:02:49
 * @Last Modified by: zengzh
 * @Last Modified time: 2023-04-23 20:21:27
 */
package dict

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	d := New()
	if d == nil {
		t.Error("New() returned nil")
	}
	if len(d.hashTables) != 2 {
		t.Error("New() did not initialize hashTables slice correctly")
	}
}
func TestString(t *testing.T) {
	d := New()
	d.hashTables[0].size = 10
	d.hashTables[0].used = 5
	str := d.String()
	expected := "Dict(len = 5, cap = 10, isRehash = false )"
	if str != expected {
		t.Errorf("String() returned %s, expected %s", str, expected)
	}
}
func TestKeyIndex(t *testing.T) {
	d := New()
	d.Store("foo", "bar")

	key := "foo"
	idx, ent := d.keyIndex(key)
	if idx != uint64(SipHash(key)&d.hashTables[0].sizemask) {
		t.Errorf("keyIndex() returned index %d, expected %d", idx, SipHash(key)&d.hashTables[0].sizemask)
	}
	if ent == nil {
		t.Error("keyIndex() returned a nil entry for an existing key")
	}

	key = "notexists"
	idx, ent = d.keyIndex(key)
	if idx != uint64(SipHash(key)&d.hashTables[0].sizemask) {
		t.Errorf("keyIndex() returned index %d for non-existing key, expected %d", idx, SipHash(key)&d.hashTables[0].sizemask)
	}
	if ent != nil {
		t.Error("keyIndex() returned a non-nil entry for a key that does not exist")
	}
}
func TestStore(t *testing.T) {
	d := New()
	key := "foo"
	value := "bar"
	d.Store(key, value)
	ent, bol := d.Load(key)
	if bol != true {
		t.Error("Store() returned ok = false for an existing key")
	}
	if ent != value {
		t.Errorf("Store() did not store the correct value for key %s", key)
	}
	d.Store(key, "newvalue")
	ent_1, bol_1 := d.Load(key)
	if bol_1 != true {
		t.Error("Store() returned ok = false for an existing key")
	}
	if ent_1 != "newvalue" {
		t.Errorf("Store() did not store the correct value for key %s", key)
	}
}
func TestLoad(t *testing.T) {
	d := New()
	key := "foo"
	value := "bar"
	d.Store(key, value)
	val, ok := d.Load(key)
	if !ok {
		t.Error("Load() returned ok = false for an existing key")
	}
	if val != value {
		t.Errorf("Load() returned value %s, expected %s", val, value)
	}
	_, ok = d.Load("notexists")
	if ok {
		t.Error("Load() returned ok = true for a non-existing key")
	}
}

func TestLoadOrStore(t *testing.T) {
	d := New()
	key := "foo"
	value := "bar"
	ent, loaded := d.loadOrStore(key, value)
	if loaded {
		t.Error("loadOrStore() returned loaded = true for a new key")
	}
	if ent != nil {
		t.Errorf("loadOrStore() did not store the correct value for new key %s", key)
	}
	ent, loaded = d.loadOrStore(key, "newvalue")
	if !loaded {
		t.Error("loadOrStore() returned loaded = false for an existing key")
	}
	if ent.value != "bar" {
		t.Errorf("loadOrStore() did not update the value for existing key %s", key)
	}
}
func TestDelete(t *testing.T) {
	d := New()
	key1, key2 := "foo", "bar"
	d.Store(key1, "val1")
	d.Store(key2, "val2")
	d.Delete(key1)
	_, ok := d.Load(key1)
	if ok {
		t.Error("Delete() did not actually delete the key")
	}
	val, _ := d.Load(key2)
	if val != "val2" {
		t.Error("Delete() deleted the wrong key")
	}
}

func TestLen(t *testing.T) {
	d := New()
	if d.Len() != 0 {
		t.Errorf("New dict has non-zero length: %d", d.Len())
	}
	d.Store("foo", "bar")
	d.Store("baz", "qux")
	if d.Len() != 2 {
		t.Errorf("Length of dict with 2 keys is %d, expected 2", d.Len())
	}
}
func TestCap(t *testing.T) {
	d := New()
	if d.Cap() != 0 {
		t.Errorf("Initial cap of dict is %d, expected 4", d.Cap())
	}
	d.Store("foo", "bar")
	d.Store("baz", "qux")
	if d.Cap() != 4 {
		t.Errorf("Cap of dict with 2 keys is %d, expected 4", d.Cap())
	}
}

func TestExpandIfNeeded(t *testing.T) {
	d := New()
	for i := 0; i < 5; i++ {
		d.Store(fmt.Sprintf("key%d", i), "val")
	}
	if d.hashTables[1].size != 8 {
		t.Errorf("Dict size after 5 inserts is %d, expected 8", d.hashTables[0].size)
	}
}
func TestResizeTo(t *testing.T) {
	d := New()
	if err := d.resizeTo(8); err != nil {
		t.Error(err)
	}
	if d.hashTables[0].size != 8 {
		t.Errorf("resizeTo(8) gave size %d, expected 8", d.hashTables[0].size)
	}
	if err := d.resizeTo(4); err != nil {
		t.Error("resizeTo(4) on dict with size 8 did not return error")
	}
	if d.hashTables[1].size != 4 {
		t.Errorf("resizeTo(4) gave size %d, expected 4", d.hashTables[0].size)
	}
}

func TestNextPower(t *testing.T) {
	d := New()
	if d.nextPower(3) != 4 {
		t.Errorf("nextPower(3) returned %d, expected 4", d.nextPower(3))
	}
	if d.nextPower(5) != 8 {
		t.Errorf("nextPower(5) returned %d, expected 8", d.nextPower(5))
	}
	if d.nextPower(8) != 8 {
		t.Errorf("nextPower(8) returned %d, expected 8", d.nextPower(8))
	}
}

func TestRehashForAWhile(t *testing.T) {
	d := New()
	for i := 0; i < 100; i++ {
		d.Store(i, i)
	}
	if d.hashTables[1].size != 128 {
		t.Errorf("Dict size is %d, expected 128", d.hashTables[1].size)
	}
	if d.isRehashing() == false {
		t.Errorf("Dict is not rehashing")
	}

	n := d.RehashForAWhile(time.Microsecond * 5)
	if d.isRehashing() == true {
		t.Errorf("Dict is rehashing")
	}
	if n != 100 {
		t.Errorf("RehashForAWhile returned %d, expected 100", n)
	}
	if d.hashTables[0].size != 64 {
		t.Errorf("Dict size after 100 rehashes is %d, expected 64", d.hashTables[0].size)
	}
}
func TestIsRehashing(t *testing.T) {
	d := New()
	if d.isRehashing() {
		t.Error("New dict isRehashing() returned true")
	}
	d.rehashIndex = 1
	if !d.isRehashing() {
		t.Error("Dict with rehashIndex set isRehashing() returned false")
	}
}

func TestIterator(t *testing.T) {
	d := New()
	d.Store("a", 1)
	d.Store("b", 2)
	d.Store("c", 3)
	it := newIterator(d, false)
	defer it.release()
	count := 0
	for ent := it.next(); ent != nil; ent = it.next() {
		count++
		key := ent.key.(string)
		value := ent.value.(int)
		if v, f := d.Load(key); v != value || f != true {
			t.Errorf("iterator returned wrong value for key %s", key)
		}
	}
	if count != 3 {
		t.Errorf("iterator iterated %d keys, expected 3", count)
	}
}
func TestIteratorRehash(t *testing.T) {
	d := New()
	d.Store("a", 1)
	d.Store("b", 2)
	it := newIterator(d, false)
	d.resizeTo(8)
	count := 0
	for ent := it.next(); ent != nil; ent = it.next() {
		count++
		key := ent.key.(string)
		value := ent.value.(int)
		if v, f := d.Load(key); v != value || f != true {
			t.Errorf("iterator returned wrong value for key %s after rehash", key)
		}
	}
	if count != 1 {
		t.Errorf("iterator iterated %d keys after rehash, expected 1", count)
	}
}
func TestRange(t *testing.T) {
	d := New()
	d.Store("a", 1)
	d.Store("b", 2)
	count := 0
	d.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("Range iterated %d keys, expected 2", count)
	}
}

func TestRangeSafely(t *testing.T) {
	d := New()
	d.Store("a", 1)
	d.Store("b", 2)
	d.RangeSafely(func(key, value interface{}) bool {
		d.Delete("a")
		return true
	})
	if v, _ := d.Load("a"); v != nil {
		t.Error("RangeSafely did not prevent deletion of iterated keys")
	}
}
