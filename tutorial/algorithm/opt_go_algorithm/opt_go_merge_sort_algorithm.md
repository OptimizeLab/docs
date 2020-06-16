# 使用二分插入技术优化Go归并排序算法的性能
### 安装包和源码准备
- [Golang源码仓库](https://go.googlesource.com/go)下载
```bash
$ git clone https://go.googlesource.com/go
$ cd go/src
```
- 硬件配置：鲲鹏(ARM64)服务器

### 1. Go归并排序函数的算法问题
Golang语言的sort包内部实现了归并排序算法，是sort对外接口的底层排序算法之一。在golang1.4发行版中，Go的归并排序算法采用递归实现归并排序的分治算法，而递归无疑增加了算法的时间和空间消耗，算法实现如下：
```go
//优化前的代码
func symMerge(data Interface, a, m, b int) {
	mid := int(uint(a+b) >> 1) // 取中间值
	n := mid + m
	var start, r int
	if m > mid {
		start = n - b
		r = mid
	} else {
		start = a
		r = m
	}
	p := n - 1

	for start < r {
		c := int(uint(start+r) >> 1)
		if !data.Less(p-c, c) {
			start = c + 1
		} else {
			r = c
		}
	}

	end := n - start
	if start < m && m < end {
		rotate(data, start, m, end)
	}
	if a < start && start < mid {
		symMerge(data, a, start, mid) // 递归处理子序列
	}
	if mid < end && end < b {
		symMerge(data, mid, end, b) // 递归处理子序列
	}
}

```
-  [归并排序算法介绍](https://zh.wikipedia.org/wiki/%E5%BD%92%E5%B9%B6%E6%8E%92%E5%BA%8F)

### 2. 使用二分插入技术优化Go归并排序函数的算法问题
#### 2.1 问题分析
通过分析go的归并排序算法，发现可以针对一些特殊的切片长度做性能优化。如果能使用性能更好的排序算法处理一些特殊长度的切片，对于symMerge函数的性能也能有提高。比如说当切片长度为1时，采用二分插入排序可以加快排序的速度。
- [二分插入算法介绍](https://baike.baidu.com/item/%E4%BA%8C%E5%88%86%E6%B3%95%E6%8F%92%E5%85%A5%E6%8E%92%E5%BA%8F)
#### 2.2 优化方案
分析golang src/sort/sort.go文件源码，发现使用二分插入排序在一定程度上优化symMerge函数的算法问题，只需要在symMerge函数中增加切片长度为1的插入排序实现即可。  
在Golang社区发行版1.5之后中已经对symMerge函数进行了优化，修复了本文提到了归并排序算法的问题，具体的CL：[sort: optimize symMerge performance for blocks with one element](https://go-review.googlesource.com/c/go/+/2219)，该优化方法在symMerge函数开头，使用二分插入法对任一切片长度为1的两个切片排序，时间复杂度为 log<sub>2</sub>N，避免了使用归并排序（时间复杂度n* log<sub>2</sub>N）带来的性能消耗。
#### 2.3 优化前后对比
![image](images/cl-2219-optCompare.PNG)
#### 2.4 优化后代码解读  
   data[a:m]和data[m:b]是两个待归并的有序切边，如果 `m-a==1`，表示data[a:m]只有一个元素，使用二分查找插入排序将元素data[a] 插入有序的data[m:b]切片；如果 `b-m==1`，表示data[m:b]只有一个元素，使用二分查找插入排序将元素data[b-1] 插入有序的data[a:m]切片。  
优化后，symMerge函数在切片长度为1时，时间复杂度为 log<sub>2</sub>N，且均耗时减小性能提升。
```go
// 优化后的代码
func symMerge(data Interface, a, m, b int) {
    // Avoid unnecessary recursions of symMerge
    // by direct insertion of data[a] into data[m:b]
    // if data[a:m] only contains one element.
    if m-a == 1 { // 切片data[a:m]长度为1
        // Use binary search to find the lowest index i
        // such that data[i] >= data[a] for m <= i < b.
        // Exit the search loop with i == b in case no such index exists.
        i := m
        j := b
        for i < j { // 二分查找合适的插入点
            h := int(uint(i+j) >> 1) 
            if data.Less(h, a) {
                i = h + 1
            } else {
                j = h
            }
        }
        // Swap values until data[a] reaches the position before i.
        for k := a; k < i-1; k++ {
            data.Swap(k, k+1) // 插入元素data[a]
        }
        return
    }

    // Avoid unnecessary recursions of symMerge
    // by direct insertion of data[m] into data[a:m]
    // if data[m:b] only contains one element.
    if b-m == 1 { // 切片data[m+1:b]长度为1
        // Use binary search to find the lowest index i
        // such that data[i] > data[m] for a <= i < m.
        // Exit the search loop with i == m in case no such index exists.
        i := a
        j := m
        for i < j { // 二分查找合适的插入点
            h := int(uint(i+j) >> 1)
            if !data.Less(m, h) {
                i = h + 1
            } else {
                j = h
            }
        }
        // Swap values until data[m] reaches the position i.
        for k := m; k > i; k-- {
            data.Swap(k, k-1) // 插入元素data[m]
        }
        return
    } 
    ...... // 归并排序的递归实现
}
```
### 3. 性能验证
使用benchstat进行性能对比，整理到表格后如下： 

测试项|优化前性能|优化后性能|性能提升
---|---|---|---|
BenchmarkStableString1K-8 | 302278 ns/op | 288879 ns/op | 4.43%
BenchmarkStableInt1K-8 | 144207 ns/op | 139911 ns/op | 2.97%
BenchmarkStableInt1K_Slice-8 | 128033 ns/op | 127660 ns/op | 0.29%
BenchmarkStableInt64K-8 | 12291195 ns/op | 12119536 ns/op | 1.40%
BenchmarkStable1e2-8 | 135357 ns/op | 124875 ns/op | 7.74%
BenchmarkStable1e4-8 | 43507732 ns/op | 40183173 ns/op | 7.64%
BenchmarkStable1e6-8 | 9005038733 ns/op | 8440007994 ns/op | 6.27%

使用二分插入技术优化归并算法后，测试用例的排序性能普遍得到提升。