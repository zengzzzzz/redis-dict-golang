package dict

import (
	"fmt"
	"github.com/dchest/siphash"
	"reflect"
)

var (
	siph = siphash.New([]byte(""))
)

func SipHash(v interface{}) uint64 {
	var data []byte
	switch iv := v.(type) {
	case string:
		data = []byte(iv)
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		data = []byte(fmt.Sprintf("%d", iv))
	default:
		panic((fmt.Sprintf("key type %s is not supported", reflect.TypeOf(iv).String())))
	}
	siph.Reset()
	_, _ = siph.Write(data)
	return siph.Sum64()
}
