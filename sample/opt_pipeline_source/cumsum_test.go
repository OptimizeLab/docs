package cumsum

import (
	"fmt"
	"testing"
)

var kArr = make([]int, 1000)

func init() {
	for i := 0; i < 1000; i++ {
		kArr[i] = i
	}
}

func TestCumsum(t *testing.T) {
	for _, v := range []int{0, 1, 7, 8, 15, 16, 127, 4095, 99999} {
		arr := generateArr(v)
		a := Cumsum(arr)
		b := CumsumChunk8(arr)
		if a != b {
			t.Errorf("results not equal, a: %d, b: %d", a, b)
			t.Error("resource: ", arr)
			break
		}
	}
}

func BenchmarkCumsum(b *testing.B) {
	bCumsum(b, Cumsum)
	//bCumsum(b, CumsumChunk8)
}

var count int

func bCumsum(b *testing.B, sum func([]int) int) {
	for _, v := range []int{0, 1, 7, 8, 15, 16, 127, 4095, 99999} {
		b.StopTimer()
		arr := generateArr(v)
		b.Run(fmt.Sprintf("%d", v), func(b *testing.B) {
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				count = sum(arr)
			}
		})
	}
}

func generateArr(n int) []int {
	arr := make([]int, n)
	from := 0
	for n > 1000 {
		copy(arr[from:from+1000], kArr[:])
		from += 1000
		n -= 1000
	}

	if n < 1000 {
		copy(arr[from:from+n], kArr[:n])
	}
	return arr
}
