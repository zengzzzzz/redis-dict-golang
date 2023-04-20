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
			in:   "string key",
			want: 10057810252675432601,
		},
		{
			name: "int input",
			in:   100,
			want: 8497251319755255498,
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
				if r := recover(); r != nil {
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
