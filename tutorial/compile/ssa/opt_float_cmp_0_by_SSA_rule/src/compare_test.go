package main

import (
	"testing"
)

func BenchmarkFloatCompare(b *testing.B) {
	arr := make([]int, 10)
	for i := 0; i < b.N; i++ {
		comp(232323.232323, arr)
	}
}
