# 基于ARM SIMD技术优化字节数组比较性能
[SIMD即单指令多数据流(Single Instruction Multiple Data)](https://en.wikipedia.org/wiki/SIMD)，通过一条指令同时对多个数据进行运算，能够有效提高CPU的运算速度，主要适用于计算密集型、数据相关性小的多媒体、数学计算、人工智能等领域。  
Golang是一个在云原生领域广泛使用的编程语言，被誉为云时代的C语言，Go在设计时充分考虑了简洁性和性能，他既包含强大的运行时，帮助用户进行协程调度、垃圾回收，也包括丰富的基础库如数学库、字节数组库、加解密库、图像库、编解码等等。对于性能要求较高且编译器目前还不能优化的场景，Go语言通过在底层使用汇编技术进行了优化，其中最重要的就是SIMD技术，下面将以一个字节数组比较的SIMD优化来详细介绍。
### 1. 安装包和源码准备
- 硬件配置：鲲鹏(ARM64)云Linux服务器[通用计算增强型KC1 kc1.2xlarge.2(8核|16GB)](https://www.huaweicloud.com/product/ecs.html)
- [Golang发行版 1.9.2 和 1.12.1](https://golang.org/dl/)，此处开发环境准备请参考文章：[Golang 在ARM64开发环境配置](https://github.com/OptimizeLab/docs/blob/master/tutorial/environment/go_dev_env/go_dev_env.md)
- [Golang github源码仓库](https://github.com/golang/go)下载，此处可以直接下载打包文件，但更好的方式是通过git工具管理
- [Git使用简介](https://www.liaoxuefeng.com/wiki/896043488029600/896067008724000)：可以参考廖雪峰老师的网站  
通过在bash命令行执行如下指令拉取golang的最新代码：
```bash
$ git clone https://github.com/golang/go
```
### 2. byte数组比较的性能问题
在代码编写中经常会碰到比较两个字符串是否相等的情况，在Golang语言中可以将两个string转为字节数组[]byte的形式进行比较，其中每个byte是一个字节(8bit)，如下是一个简单直观的比较函数实现思路:
```go
func EqualByteArrAB(a, b []byte) bool {
    //---------------数组长度比较-------------------
    if len(a) != len(b) {          
        return false               
    }                              
    //--------按顺序比较数组a和数组b中的每个byte------
    for i, _ := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}
```
实现算法后，需要进行充分的测试才能投入生产环境，通过[benchmark工具](https://golang.org/pkg/testing/)进行测试后获得如下性能数据：
```bash
goos: linux
goarch: arm64
pkg: test_obj/testbyteequal
BenchmarkEqual/0-8             330669548                3.64 ns/op
BenchmarkEqual/1-8             227632882                5.27 ns/op           189.74 MB/s
BenchmarkEqual/6-8             132229749                9.09 ns/op           660.35 MB/s
BenchmarkEqual/9-8             100000000                10.1 ns/op           893.80 MB/s
BenchmarkEqual/15-8             83173801                14.4 ns/op          1041.32 MB/s
BenchmarkEqual/16-8             79955283                15.0 ns/op          1069.79 MB/s
BenchmarkEqual/20-8             67353938                17.8 ns/op          1124.26 MB/s
BenchmarkEqual/32-8             45706566                26.2 ns/op          1219.49 MB/s
BenchmarkEqual/4K-8               421956                2844 ns/op          1440.18 MB/s
BenchmarkEqual/4M-8                  334             3496666 ns/op          1199.52 MB/s
BenchmarkEqual/64M-8                  18            66481026 ns/op          1009.44 MB/s
```
通过观察数据可以直观的发现，随着数据增长，算法的性能变得非常差，当字节数组长度>4K时，已经影响到了正常的使用。  

对代码进行简单分析，可以发现最主要的操作是循环执行比较语句a[i] != b[i]，这里最主要的开销是byte数据加载和比较操作，随着测试的byte数组长度增长，执行的数据加载和比较操作越来越多，开销也就越来越大。如果能够减少数据加载和比较的次数就可以优化性能，利用ARM硬件架构的SIMD加速特性是解决该问题的有效办法：
### 3. SIMD优化方案
那么Go语言是怎么处理这个问题的呢？在深入分析本文所使用的Golang发行版1.9.2官方源码后，在src/runtime/asm_arm64.s文件中发现了上节比较函数的Golang汇编版本EqualBytes，算法逻辑完全一致，如图所示：
![image](images/image-code-compare.png)
经过对比验证发现在常用的社区发行版(如1.12.1)中已经对EqualBytes函数做了优化，解决了本文场景下的性能问题，具体的优化提交记录见[ChangeList](http://svnbook.red-bean.com/en/1.8/svn.advanced.changelists.html)
：[bytes: add optimized Equal for arm64](https://go-review.googlesource.com/c/go/+/71110)，该优化方法使用了SIMD技术，即单指令同时处理多个byte数据，大幅减少了数据加载和比较操作的指令条数，提升了性能
### 4. SIMD优化前后对比  
如下是使用SIMD技术优化前后的对比图，优化前后的代码都是使用Golang汇编编写，从图中可以看到优化前代码非常简单，循环取1 byte进行比较，使用SIMD指令优化后，代码变得非常复杂，这里可以先避免陷入细节，先理解实现原理，具体代码细节可以在章节6再进一步学习。此处代码变复杂的主要原因是进行了分情况的分块处理，首先循环处理64 bytes大小的分块，当数组末尾不足64 bytes时，再将余下的按16 bytes分块处理，直到余下长度为1时的情况，下图直观的演示了优化前后的对比关系和优化后分块处理的规则:
![image](images/simd-compare.png)  
    
### 5. 优化前代码详解
优化前的代码实现在src/runtime/asm_arm64.s中，该函数是循环从两个数组中取1 byte进行比较，每byte数据要耗费两个加载操作、1个byte比较操作、1个数组末尾判断操作，如下所示：
```assembly
//func Equal(a, b []byte) bool
TEXT bytes·Equal(SB),NOSPLIT,$0-49
//---------数据加载------------
    // 将栈上数据取到寄存器中
    // 对数组长度进行比较，如果不相等直接返回0
    MOVD a_len+8(FP), R1      // 取数组a的长度
    MOVD b_len+32(FP), R3     // 取数组b的长度
    CMP R1, R3                // 数组长度比较
    BNE notequal              // 数组长度不同，跳到notequal
    MOVD a+0(FP), R0          // 将数组a的地址加载到通用寄存器R0中
    MOVD b+24(FP), R2         // 将数组b的地址加载到通用寄存器R2中
    ADD R0, R1                // R1保存数组a末尾的地址
//-----------------------------
//--------数组循环比较操作------- 
loop:
    CMP R0, R1                // 判断是否到了数组a末尾
    BEQ equal                 // 如果已经到了末尾，说明之前都是相等的，跳转到标签equal
    MOVBU.P 1(R0), R4         // 从数组a中取一个byte加载到通用寄存器R4中
    MOVBU.P 1(R2), R5         // 从数组b中取一个byte加载到通用寄存器R5中
    CMP R4, R5                // 比较寄存器R4、R5中的值
    BEQ loop                  // 相等则继续下一轮循环操作
//----------------------------- 
//-------------不相等-----------
notequal:
    MOVB ZR, ret+48(FP)       // 数组不相等，返回0
    RET
//----------------------------- 
//-------------相等------------- 
equal:
    MOVD $1, R0               // 数组相等，返回1
    MOVB R0, ret+48(FP)
    RET
//----------------------------- 
```

### 6. 优化后代码详解
优化后的代码实现在src/runtime/asm_arm64.s中。这里的代码实现因为做了循环展开，所有看起来比较复杂，但逻辑是很清晰的，即采用分块的思路，将数组划分为64/16/8/4/2/1bytes大小的块，最大程度发挥SIMD指令的优势，使用多个向量寄存器，每次循环中尽可能处理更多的数据。汇编代码解读如下(代码中添加了关键指令注释）：
```assembly
// 函数的参数，此处是通过寄存器传递参数的
// 调用memeqbody的父函数已经将参数放入了如下寄存器中
// R0: 寄存器R0保存数组a的地址
// R1: 寄存器R1数组a的末尾地址
// R2: 寄存器R2保存数组b的地址
// R8: 寄存器R8存放比较的结果
TEXT runtime·memeqbody<>(SB),NOSPLIT,$0
//---------------数组长度判断-----------------
// 根据数组长度判断按照何种分块开始处理
    CMP    $1, R1                        // 数组长度为1，跳转到标签one下面的代码
    BEQ    one
    CMP    $16, R1                       // 处理数组长度小于16的情况
    BLO    tail
    BIC    $0x3f, R1, R3                 // 位清除指令，清除R1的后6位存放到R3
    CBZ    R3, chunk16                   // 跳转指令，R3为0，跳转到chunk16
    ADD    R3, R0, R6                    // R6为64byte块尾部指针

//------------处理长度为64 bytes的块-----------
// 按64 bytes为块循环处理
chunk64_loop: 
// 加载RO,R2指向的数据块到SIMD向量寄存器中，并将RO,R2指针偏移64位                          
    VLD1.P (R0), [V0.D2, V1.D2, V2.D2, V3.D2]
    VLD1.P (R2), [V4.D2, V5.D2, V6.D2, V7.D2]
// 使用SIMD比较指令，一条指令比较128位，即16个bytes，结果存入V8-v11寄存器
    VCMEQ  V0.D2, V4.D2, V8.D2           
    VCMEQ  V1.D2, V5.D2, V9.D2
    VCMEQ  V2.D2, V6.D2, V10.D2
    VCMEQ  V3.D2, V7.D2, V11.D2  
// 通过SIMD与运算指令，合并比较结果，最终保存在寄存器V8中    
    VAND   V8.B16, V9.B16, V8.B16        
    VAND   V8.B16, V10.B16, V8.B16
    VAND   V8.B16, V11.B16, V8.B16 
// 下面指令判断是否末尾还有64bytes大小的块可继续用这里循环
// 判断是否相等，不相等则直接跳到not_equal返回
    CMP    R0, R6                        // 比较指令，比较RO和R6的值，修改寄存器标志位，对应下面的BNE指令      
    VMOV   V8.D[0], R4 
    VMOV   V8.D[1], R5                   // 转移V8寄存器保存的结果数据到R4,R5寄存器
    CBZ    R4, not_equal 
    CBZ    R5, not_equal                 // 跳转指令，若R4,R5寄存器的bit位出现0，表示不相等，跳转not_equal
    BNE    chunk64_loop                  // 标志位不等于0，对应上面RO!=R6则跳转chunk64_loop
    AND    $0x3f, R1, R1                 // 仅保存R1末尾的后6位，这里保存的是末尾不足64bytes块的大小
    CBZ    R1, equal                     // R1为0,跳转equal，否则向下顺序执行

//-----------处理剩余长度小于16的块------------
chunk16:                               
    BIC    $0xf, R1, R3                  // 位清除指令，清除R1的后4位存到R3
    CBZ    R3, tail                      // R3为0，表示末尾剩余的块小于16byte，跳转到tail块
    ADD    R3, R0, R6                    // R6为16byte块尾部指针
//-----------循环处理长度为16 bytes的块------------
chunk16_loop:                            // 循环处理16byte，处理过程类似chunk64_loop
    VLD1.P (R0), [V0.D2] 
    VLD1.P (R2), [V1.D2]
    VCMEQ    V0.D2, V1.D2, V2.D2
    CMP R0, R6
    VMOV V2.D[0], R4
    VMOV V2.D[1], R5
    CBZ R4, not_equal                  // 判断是否有不等，如有0位，跳not-equal
    CBZ R5, not_equal
    BNE chunk16_loop                   // 末尾还有至少一个16bytes的大小，循环继续
    AND $0xf, R1, R1
    CBZ R1, equal                      // 若无剩余块（小于16byte），则跳转equal，否则向下顺序执行
//---------处理在末尾长度小于16 bytes的块---------
tail:                                  
    TBZ $3, R1, lt_8                   // 跳转指令，若R1[3]==0，也就是R1小于8，跳转到lt_8
    MOVD.P 8(R0), R4
    MOVD.P 8(R2), R5
    CMP R4, R5    
    BNE not_equal 
//---------处理在末尾长度小于8 bytes的块---------
lt_8:                                  
    TBZ $2, R1, lt_4
    MOVWU.P 4(R0), R4
    MOVWU.P 4(R2), R5
    CMP R4, R5
    BNE not_equal
//---------处理在末尾长度小于4 bytes的块---------
lt_4:                                 
    TBZ $1, R1, lt_2
    MOVHU.P 2(R0), R4
    MOVHU.P 2(R2), R5
    CMP R4, R5
    BNE not_equal
//---------处理在末尾长度小于2 bytes的块---------
lt_2:                                 
    TBZ     $0, R1, equal
//-----------处理在末尾长度为1 byte的块----------
one:                                  
    MOVBU (R0), R4
    MOVBU (R2), R5
    CMP R4, R5
    BNE not_equal
//-----------------判断相等返回1----------------
equal:
    MOVD $1, R0
    MOVB R0, (R8)
    RET
//----------------判断不相等返回0----------------
not_equal:
    MOVB ZR, (R8)
    RET
```
上述优化代码中，使用VLD1(数据加载指令)一次加载64byte数据到SIMD寄存器，再使用VCMEQ指令比较SIMD寄存器保存的数据内容得到结果，相比传统用的单字节比较方式，大大提高了大于64byte数据块的比较性能。大于16byte小于64byte块数据，使用一个SIMD寄存器一次处理16byte块的数据，小于16byte数据块使用通用寄存器保存数据，一次比较8\4\2\1byte的数据块。

### 7. 结果验证
使用[benchstat工具](https://godoc.org/golang.org/x/perf/cmd/benchstat)进行性能对比，整理到表格后如下所示： 
![image](images/SIMDEqualResult.png)  
上表中可以清晰的看到使用SIMD优化后，所有的用例性能都有所提升，其中数据大小为4K时性能提升率最高，耗时减少了93.48%；每秒数据处理量提升14.29倍
