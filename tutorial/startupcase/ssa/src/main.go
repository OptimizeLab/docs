package main

func comp(x float64, arr []int) {
	for i := 0; i < len(arr); i++ {
		if x > 0 {
			arr[i] = 1
		}
	}
}

func main() {
}
