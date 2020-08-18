# 基于二分插入技术优化归并排序算法
> 本文以Go语言的排序算法优化实践为例，讲解通用的优化方法。
> 基于Golang社区的优化补丁[sort: optimize symMerge performance for blocks with one element](https://go-review.googlesource.com/c/go/+/2219)，展开分析归并排序算法的优化。

### 1. 安装包和源码准备
- 硬件配置：鲲鹏(ARM64)云Linux服务器[通用计算增强型KC1 kc1.2xlarge.2(8核|16GB)](https://www.huaweicloud.com/product/ecs.html)
- [Go发行版1.4 和 1.5](https://golang.org/dl/)，此处开发环境准备请参考文章：[Golang 在ARM64开发环境配置](https://github.com/OptimizeLab/docs/blob/master/tutorial/environment/go_dev_env/go_dev_env.md)
- [Golang github源码仓库](https://github.com/golang/go)下载，此处可以直接下载打包文件，但更好的方式是通过git工具管理
- [Git使用简介](https://www.liaoxuefeng.com/wiki/896043488029600/896067008724000)：可以参考廖雪峰老师的网站  
    通过在bash命令行执行如下指令拉取golang的最新代码：  
    ```bash
    $ git clone https://github.com/golang/go
    ```

### 2. 基于Go1.4发行版的归并排序算法举例
Go语言的sort包对插入排序、归并排序、快速排序、堆排序做了支持，
[归并排序](https://zh.wikipedia.org/wiki/%E5%BD%92%E5%B9%B6%E6%8E%92%E5%BA%8F)
是sort包实现的稳定排序算法。在Go1.4发行版中，使用递归实现了归并排序算法，代码如下：
```go
//优化前的归并排序代码，使用递归实现两个子序列data[a：m]和data [m：b]合并 
func symMerge(data Interface, a, m, b int) {
	// 计算数据块a:b的中位数mid和a:m数据块中用于数据交换的起始下标start和r
	mid := int(uint(a+b) >> 1) 
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

	// 折半查找出a:m数据块中适合的区间start:m，用于后续的数据交换
	for start < r {
		c := int(uint(start+r) >> 1)
		if !data.Less(p-c, c) {
			start = c + 1
		} else {
			r = c
		}
	}

	// 旋转交换数据块 start:m 和 m:end 
	end := n - start
	if start < m && m < end {
		rotate(data, start, m, end)
	}
	// 旋转交换后的两组数据段 a:start和start:mid，mid:end和end:b分别进行归并排序
	if a < start && start < mid {
		symMerge(data, a, start, mid)
	}
	if mid < end && end < b {
		symMerge(data, mid, end, b) 
	}
}
```

### 3. 问题分析
Go1.4基于递归实现了归并排序算法，是常用的归并排序算法实现方法，也是算法的性能损耗的主要原因之一。
如果可以在一些应用场景减少对递归函数的调用，算法的性能可以得到提升。比如说当处理切片长度为1的归并时，
使用[插入排序](https://baike.baidu.com/item/%E4%BA%8C%E5%88%86%E6%B3%95%E6%8F%92%E5%85%A5%E6%8E%92%E5%BA%8F)性能会明显好于归并排序，
而这种场景也是常见的，在不断递归处理子切片的过程中，出现切片长度为1的情况也是常有的。

### 4. 优化方案
通过分析Go1.4归并排序的代码和问题，找到了一种优化方法。即在归并排序中，
针对切片长度为1的数据块，使用插入排序可以减少函数递归的调用，提高算法的性能。     
在Go1.5中已经应用了这种优化方法，提升了归并排序的性能，优化补丁：
[sort: optimize symMerge performance for blocks with one element](https://go-review.googlesource.com/c/go/+/2219)。
使用二分插入排序算法处理切片长度为1的排序，插入排序时间复杂度为 log<sub>2</sub>N，避免了归并排序算法调用递归带来的性能消耗，归并排序时间复杂度n* log<sub>2</sub>N。

### 5. 优化实现
优化前后的代码对比  
![image](images/cl-2219-optCompare.PNG)

优化代码分析如下：
```go
// 优化后的代码，基于二分插入排序提升的归并排序算法性能
func symMerge(data Interface, a, m, b int) {
    // data[a:m]只有一个元素时，使用二分插入排序直接插入data[a]到data[m:b] 
    // 可以避免调用递归函数，提高性能
    if m-a == 1 { 
        i := m
        j := b
        for i < j {
            h := int(uint(i+j) >> 1) 
            if data.Less(h, a) {
                i = h + 1
            } else {
                j = h
            }
        }
        for k := a; k < i-1; k++ {
            data.Swap(k, k+1)
        }
        return
    }

    // data[m:b]只有一个元素时，使用二分插入排序直接插入data[m]到data[a:m]
    // 可以避免调用递归函数，提高性能
    if b-m == 1 { 
        i := a
        j := m
        for i < j { 
            h := int(uint(i+j) >> 1)
            if !data.Less(m, h) {
                i = h + 1
            } else {
                j = h
            }
        }
        for k := m; k > i; k-- {
            data.Swap(k, k-1)
        }
        return
    }
    
    // 归并排序的递归实现
    ......
}
```

### 6. 性能验证
使用[benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat)进行性能对比，整理到表格后如下： 

测试项|优化前性能|优化后性能|性能提升
---|---|---|---|
BenchmarkStableString1K-8 | 302278 ns/op | 288879 ns/op | 4.43%
BenchmarkStableInt1K-8 | 144207 ns/op | 139911 ns/op | 2.97%
BenchmarkStableInt1K_Slice-8 | 128033 ns/op | 127660 ns/op | 0.29%
BenchmarkStableInt64K-8 | 12291195 ns/op | 12119536 ns/op | 1.40%
BenchmarkStable1e2-8 | 135357 ns/op | 124875 ns/op | 7.74%
BenchmarkStable1e4-8 | 43507732 ns/op | 40183173 ns/op | 7.64%
BenchmarkStable1e6-8 | 9005038733 ns/op | 8440007994 ns/op | 6.27%

[注] `BenchmarkStableString1K`表示长度为1K的字符串切片，`-8`表示运行时的GOMAXPROCS的值，`BenchmarkStableInt1K`表示长度为1K的整型切片，`BenchmarkStable1e2-8`表示长度为1e2(100)的结构体切片。

性能测试结果显示，使用二分插入技术优化归并排序算法后，算法的性能整体得到提升。

### 7. 总结
这个优化案例从排序算法的特定应用场景出发，找到了一种优化方法，并提升了算法的整体性能，提供了一种值得学习借鉴的算法优化思路。