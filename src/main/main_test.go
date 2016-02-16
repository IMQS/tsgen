package main

import (
	"testing"
)

func BenchmarkMain(b *testing.T) {
	for n := 0; n < b.N; n++ {
		initilaise()
	}

}
