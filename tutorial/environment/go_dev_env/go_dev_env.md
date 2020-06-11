# Golang 在ARM64开发环境配置

### 1. 在ARM64服务器上配置开发环境

Go语言开发包是go语言的实现，内容包括版本的语法、编译、运行、标准库以及其他一些必要资源。

1） Goland语言包官网：http://goland.org/dl/。

2） 下载打开页面之后根据自己的需求选择对应平台下载，本次安装选择[go1.14.2.linux-arm64](https://dl.google.com/go/go1.14.2.linux-arm64.tar.gz) ,是发行版1.14.2对应linux系统、arm64处理器架构的特定版本。如下图所示：

![image](images/image-version.png)

3） 进入你的linux平台，进入你存放安装包的目录下，输入命令：

```linux
wget https://dl.google.com/go/go1.14.2.linux-arm64.tar.gz
```

进行下载，下载结果如图所示。

![image](images/image-install.png)

4） 执行tar解压到/usr/loacl目录下（官方推荐)，得到go文件夹。

```linux
tar -C /usr/loacl -zxvf go1.14.2.linux-arm64.tar.gz
```

得到go文件夹内容，结果如图所示：

![image](images/image-gofile.png)

5） 配置环境变量，输入命令：

```linux
export GOROOT=/usr/loacl/go
export PATH=$PATH:$GOROOT/bin
```

6） 输入以下命令就可以得到你的版本号：

```linux
go version
```

7） 新建一个工作目录并且创建第一个工程目录：

```linux
#创建工作空间
mkdir $HOME/go
#编辑 ~/.bash_profile 文件
#将你的工作目录声明到环境变量中
export GOPATH=$HOME/go
#保存退出后source一下
source ~./bash_profile
#之后创建并进入你的第一个目录
mkdir -p $GOPATH/hello && cd $GOPATH/src/hello
```

8） 在工作目录下创建名为hello.go 的文件。内容如下：

```go
package main

import "fmt"

func main() {
	fmt.Printf("hello, world\n")
}
```

9） 使用命令：go build hello.go，来构建然后使用命令：./hello来运行。

![image](images/image-result.png)

10） 到这里Golang开发环境就准备完毕了

### 2. Golang官方仓库准备

Golang是一个开源的项目，每个人都可以贡献代码，下面我们讲解如何获取官方仓库
```bash
$ git clone https://go.googlesource.com/go
$ cd go/src
```
根据Commit ID可以自由创建分支进行测试，如下通过ID: 0c68b79创建了名称为test-simd的分支
```bash
$ git checkout -b test-simd 0c68b79
```
从源码编译并测试Golang
```bash
$ cd go/src
$ bash all.bash 
```
编译后我们将在go/bin目录下获得go工具

### 3 使用benchmark获取性能分析报告

设置临时环境变量（仅在本条命令中有效），并使用新编译的 Go 执行测试用例，当然这里也可以直接使用第一节中安装的go发布版来执行测试用例
```bash
$ GOROOT=`pwd`/..; $GOROOT/bin/go test bytes -v -bench ^BenchmarkEqual$ -run ^$ -cpuprofile=cpu.out
goos: linux
goarch: arm64
pkg: bytes
BenchmarkEqual/0-8              500000000               3.84 ns/op
BenchmarkEqual/1-8              300000000               5.44 ns/op       183.74 MB/s
BenchmarkEqual/6-8              100000000               10.0 ns/op       598.26 MB/s
BenchmarkEqual/9-8              100000000               12.4 ns/op       728.51 MB/s
BenchmarkEqual/15-8             100000000               17.0 ns/op       880.77 MB/s
BenchmarkEqual/16-8             100000000               17.8 ns/op       901.06 MB/s
BenchmarkEqual/20-8             100000000               20.9 ns/op       956.05 MB/s
BenchmarkEqual/32-8              50000000               30.5 ns/op      1050.11 MB/s
BenchmarkEqual/4K-8                500000               3176 ns/op      1289.51 MB/s
BenchmarkEqual/4M-8                   500            3468668 ns/op      1209.20 MB/s
BenchmarkEqual/64M-8                   20           53439570 ns/op      1255.79 MB/s
PASS
ok      bytes   18.907s
```
结果被保存在 cpu.out 文件中

#### 4  使用pprof工具进一步分析性能
通过 go 提供的性能分析工具 pprof 操作 cpu.out 文件，我们可以轻松查看并分析对 cpu 性能消耗大的函数。
```bash
$ go tool pprof cpu.out
File: bytes.test
Type: cpu
Time: Apr 16, 2020 at 6:20pm (CST)
Duration: 19.97s, Total samples = 19.93s (99.81%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top 3
Showing nodes accounting for 19.75s, 99.10% of 19.93s total
Dropped 14 nodes (cum <= 0.10s)
Showing top 3 nodes out of 8
      flat  flat%   sum%        cum   cum%
    14.32s 71.85% 71.85%     14.32s 71.85%  bytes.Equal /home/chan/go/src/runtime/asm_arm64.s
     4.06s 20.37% 92.22%     17.49s 87.76%  bytes_test.bmEqual.func1 /home/chan/go/src/bytes/bytes_test.go
     1.37s  6.87% 99.10%      2.31s 11.59%  bytes_test.BenchmarkEqual.func1 /home/chan/go/src/bytes/bytes_test.go
```
此处 " top 3 " 列出了 cpu 消耗前 3 的函数。其中各项含义如下：
- flat：当前函数占用CPU的耗时  
- flat%: 当前函数占用CPU的耗时百分比  
- sun%：函数占用CPU的耗时累计百分比  
- cum：当前函数加上调用当前函数的函数占用CPU的总耗时  
- cum%：当前函数加上调用当前函数的函数占用CPU的总耗时百分比  
- 最后一列：函数名称  

我们通过 pprof 的 list 命令可以查看 bytes.Equal 方法内部的详细耗时信息：
```bash
(pprof) list bytes.Equal
Total: 19.93s
ROUTINE ======================== bytes.Equal in /home/xxx/go/src/runtime/asm_arm64.s
    14.32s     14.32s (flat, cum) 71.85% of Total
         .          .    865:   MOVD    R0, ret+24(FP)
         .          .    866:   RET
         .          .    867:
         .          .    868:// TODO: share code with memequal?
         .          .    869:TEXT bytes·Equal(SB),NOSPLIT,$0-49
     140ms      140ms    870:   MOVD    a_len+8(FP), R1
     410ms      410ms    871:   MOVD    b_len+32(FP), R3
      70ms       70ms    872:   CMP     R1, R3          // unequal lengths are not equal
         .          .    873:   BNE     notequal
     490ms      490ms    874:   MOVD    a+0(FP), R0
     300ms      300ms    875:   MOVD    b+24(FP), R2
         .          .    876:   ADD     R0, R1          // end
         .          .    877:loop:
     2.63s      2.63s    878:   CMP     R0, R1
      10ms       10ms    879:   BEQ     equal           // reaches the end
     5.72s      5.72s    880:   MOVBU.P 1(R0), R4
        4s         4s    881:   MOVBU.P 1(R2), R5
         .          .    882:   CMP     R4, R5
      40ms       40ms    883:   BEQ     loop
         .          .    884:notequal:
         .          .    885:   MOVB    ZR, ret+48(FP)
         .          .    886:   RET
         .          .    887:equal:
     380ms      380ms    888:   MOVD    $1, R0
     130ms      130ms    889:   MOVB    R0, ret+48(FP)
         .          .    890:   RET
         .          .    891:
         .          .    892:TEXT runtime·return0(SB), NOSPLIT, $0
         .          .    893:   MOVW    $0, R0
         .          .    894:   RET
```
上面结果展示的是 bytes.Equal 汇编代码的耗时情况，该汇编代码在 runtime 包的 asm_arm64.s 文件中，可以看到主要的耗时热点是 CMP 指令和 MOVBU.P 指令

### 5. 使用benchstat对比优化前后的性能数据
有时我们做了一个优化，希望比较下优化前后两组benchmark数据，这时候我们可以使用benchstat，
```bash
benchstat before-optimize after-optimize
```
获得类似如下的性能对比结果
![image](images/SIMDEqualResult.png)  
