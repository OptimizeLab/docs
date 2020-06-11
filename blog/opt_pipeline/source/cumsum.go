package cumsum

// Cumsum Returns the cumulative sum of all elements in the slice
func Cumsum(arr []int) (count int) {
	for i := 0; i < len(arr); i++ {
		count += arr[i]
	}
	return
}

// CumsumChunk8 Returns the cumulative sum of all elements in the slice, added by chunk
func CumsumChunk8(arr []int) (count int) {
	if len(arr) == 0 {
		return
	}

	// len >= 8 时，每 8 个元素一次循环，分成 4 组相加，减少依赖；
	// 使用 4 个变量暂存计算结果，促使更多的寄存器被使用
	for len(arr) >= 8 {
		a := arr[0] + arr[1]
		b := arr[2] + arr[3]
		c := arr[4] + arr[5]
		d := arr[6] + arr[7]
		a += c
		b += d
		count += a + b
		arr = arr[8:]
	}

	// 4 <= len < 8 ，取 4 个元素，分成 2 组相加
	if len(arr) >= 4 {
		a := arr[0] + arr[1]
		b := arr[2] + arr[3]
		count += a + b
		arr = arr[4:]
	}

	// 2 <= len < 4, 取 2 个元素
	if len(arr) >= 2 {
		count += arr[0] + arr[1]
		arr = arr[2:]
	}

	// len == 1, 直接相加
	if len(arr) == 1 {
		count += arr[0]
	}
	return
}
