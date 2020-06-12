# 使用并行化技术优化UTF8验证算法的性能
### 安装包和源码准备
- [Golang发行版 1.14 && 1.14.3](https://golang.org/dl/)
- [Golang源码仓库](https://go.googlesource.com/go)下载
```bash
$ git clone https://go.googlesource.com/go
$ cd go/src
```
- 硬件配置：鲲鹏(ARM64)服务器

### 1. UTF8验证函数的算法问题
Golang语言支持对UTF8编码进行验证，在使用编码校验时常使用UTF8编码验证算法对数据编码做验证。而ASCII编码检查也是使用UTF8验证函数，golang1.14发行版中的UTF8验证算法如下。每次仅验证一个byte字符是否属于ASCII编码。这导当致数据为纯ASCII编码或连续ASCII编码时，验证性能低下。
```go 
// 优化前的utf8验证算法
func Valid(p []byte) bool {
	n := len(p)
	for i := 0; i < n; {
		pi := p[i]
		if pi < RuneSelf { // 每次验证一个byte字符是否属于ASCII编码
			i++
			continue
		}
.......//验证byte字符是否属于UTF8编码
    }
}
```
### 2. 使用并行化验证优化UTF8验证函数的算法问题
#### 2.1 问题分析
通过分析现有的utf8算法，发现问题主要在于每次只检查一个byte字符是否属于ASCII编码，如果能够一次就比较多个byte字符是否属于ASCII编码，就能加速验证效率。而ASCII编码特点也适用于并行化验证。
- [UTF8编码介绍](https://zh.wikipedia.org/wiki/UTF-8)
#### 2.2 优化方案
分析golang src/unicode/utf8/utf8.go文件源码，发现使用并行化验证ASCII编码优化UTF8验证函数是可行的，只需要在utf8函数中增加ASCII编码并行化验证即可。  
在社区发行版1.14.3中已经对Valid函数进行了优化，修复了本文提到了utf8验证算法问题，具体的CL：[ unicode/utf8: optimize Valid and ValidString for ASCII checks ](https://go-review.googlesource.com/c/go/+/228823)，该优化方法在Valid函数开头，先做了ASCII编码并行化验证，一次检查8个ASCII字符，加快了ASCII编码的验证速度。
#### 2.3 优化前后对比
![image](images/cl-228823-optCompare.PNG)
#### 2.4 优化后代码解读
Valid函数首先检查ASCII编码，每次检查8个byte字符是否为ASCII，如果是循环检查；如果不是跳出循环，执行下面的UTF8编码检查。
循环内，加载byte数组的8个byte到两个unit32中。ASCII编码特点：每个字符占用8个bit位，且最高位为0，任一个ASCII编码和`0x80`的与操作结果为0。所以代码`(first32|second32)&0x80808080`可以一次同时检查8个byte是否为ASCII编码。
```go
// 优化后的utf8验证代码
func Valid(p []byte) bool {
	// Fast path. Check for and skip 8 bytes of ASCII characters per iteration.
	for len(p) >= 8 {
		// Combining two 32 bit loads allows the same code to be used
		// for 32 and 64 bit platforms.
		// The compiler can generate a 32bit load for first32 and second32
		// on many platforms. See test/codegen/memcombine.go.
		first32 := uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16 | uint32(p[3])<<24
		second32 := uint32(p[4]) | uint32(p[5])<<8 | uint32(p[6])<<16 | uint32(p[7])<<24
		if (first32|second32)&0x80808080 != 0 { // 一次检查8个ASCII编码
			// Found a non ASCII byte (>= RuneSelf).
			break
		}
		p = p[8:]
	}
	......// 验证UTF8编码
}
```
### 3. 结果验证
使用benchstat进行性能对比，整理到表格后如下： 

测试项|优化前性能|优化后性能|性能提升
---|---|---|---
BenchmarkValidTenASCIIChars-8 | 15.8 ns/op | 8.00 ns/op | 49.37%
BenchmarkValidStringTenASCIIChars-8 | 12.8 ns/op | 8.04 ns/op | 37.19%

优化后，ASCII编码的测试用例性能提升明显，性能最高提升了49%。
   
