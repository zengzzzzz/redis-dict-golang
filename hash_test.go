package dict

import (
	"testing"
    "fmt"
)

func TestSipHash(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want uint64
	}{
		{
			name: "string input",
			in:   "hello world",
			want: 5333013848549256545,
		},
		{
			name: "int input",
			in:   123,
			want: 15493594227354133703,
		},
		{
			name: "unsupported input type",
			in:   []int{1, 2, 3},
			want: 0, // expected panic
		},
	}
    fmt.Println(tests)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					if tt.want != 0 {
						t.Errorf("SipHash() did not panic on unsupported input type")
					}
				}
			}()

			got := SipHash(tt.in)

			if tt.want != got {
				t.Errorf("SipHash() = %d, want %d", got, tt.want)
			}
		})
	}
}
