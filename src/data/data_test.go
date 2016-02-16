package data

import (
	"testing"
)

func BenchmarkInit(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Process()
	}
}
