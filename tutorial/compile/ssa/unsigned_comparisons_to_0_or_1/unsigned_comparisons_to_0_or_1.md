# 通过优化无符号整数与0/1对比提升软件性能
> 本文基于分析go社区在[ARM64无符号整型数对0/1比较](https://go-review.googlesource.com/c/go/+/246857)的优化方案，使用更好的指令组合，提升软件性能。

## 1.无符号整型数与0/1对比的特殊性

往往我们写代码的时候有一个这样的思维逻辑，举个例子：
```go
var array = [5]uint8{0,1,2,3,4}
```
如上面的数组，我们需要挑出大于0的数时，往往会写成这样。
```go
for i := 0; i<len(array); i++ {
    if array[i] > 0 {
        ...
    }
}
```
由于无整型数最小值为0，这个判断条件可以改为这样以下代码，可以实现同样的功能。
```go
if array[i] != 0 {
    ...
}
```
## 2.为什么可以做优化
我以unit8举例子，unit16,unit32,unit64一样适用。
```go
func UintComp(a uint8) int {
	if a > 0 {
		return 1
	}
	return 0
}
```
通过GOSSAFUNC查看genssa编译阶段生成的指令。
![images](images/normal_comp.png)
可以从产生出来的ssa.html观测使用了什么指令
```
MOVBU "".a(RSP), R0 # MOVBU指令是将a的值读到寄存器R0中
CMPW $0, R0         # 通过CMPW指令将1和R0放到一个逻辑运算寄存器中
BLS 9               # BLS指令是用做无符号整型数小于等于对比，相当于a<=0时执行到genssa中的第9行
                    # 相当于不执行if里面的代码。
```                 
根据我们在第一段分析的结果，可以改写成这样
```go
func UintComp(a uint8) int {
	if a ！= 0 {
		return 1
	}
	return 0
}
```
重新使用GOSSAFUNC查看genssa编译阶段生成的指令,可以清楚观察到少使用了一个指令。
![images](images/arm64_CBNZ_test.png)
```
MOVBU "".a(RSP), R0
CBZW R0, 8     
# CBZW指令是将R0寄存器中的数，直接做 R0!=0 这个判断，若判断是错误的，则执行的genssa中的第8行。
# 即直接执行return 0
```
由于涉及CMPW,BLS,CBZW等指令，为了方便理解源代码，我这里引用[ARM64V8](https://developer.arm.com/documentation/den0024/a/The-A64-instruction-set/Data-processing-instructions/Conditional-instructions?lang=en)官方文档稍做解释。

在ssa/gen/ARM64Ops.go中，有如下代码:
```go
blocks := []blockData{
	...
	{name: "LE", controls: 1},	
	//L代表less，即小于，E代表英文equal,即等于，表达出来的就是当 <= 
	...
	{name: "ULE", controls: 1}, 
	//而ULE比LE多了个U，代表的是无符号 <=,相当于上面看到的指令BLS，因为B是跳转指令，可以解读为当符合这个条件时跳转。这里的名字和官方手册中有所不同，这里把官方中的LS用ULE来替换了，我认为的是更加方便理解。
	...             
	{name: "ZW", controls: 1},                 
	//上面用到的CBZW可以这么拆分 C B ZW，C是对比，B是跳转，ZW就是32位的变量==0，
	{name: "NZW", controls: 1},     
	//NZW比ZW多了个N，就是NOT，32位的变量!=0        
	...
}	
```
这些代码中
## 3.通过改写编译规则做优化
经过文章开头的分析得出，我们可以优化下列4种判断情况
```
0 <  x  =>  x != 0
x <= 0  =>  x == 0
x <  1  =>  x == 0
1 <= x  =>  x != 0
```
通过在cmd/internal/ssa/gen/ARM64.rules中加入新的SSA编译规则，根据上面表达式进行添加代码。
```
(Less(8U|16U|32U|64U) zero:(MOVDconst [0]) x) => (Neq(8|16|32|64) zero x)
 //0是否小于无符号整型x => 无符号整型x是否不等于0
(Leq(8U|16U|32U|64U) x zero:(MOVDconst [0]))  => (Eq(8|16|32|64) x zero)
 //无符号整型x是否小于或等于0 => 无符号整型x是否等于0
(Less(8U|16U|32U|64U) x (MOVDconst [1])) => (Eq(8|16|32|64) x (MOVDconst [0]))
 //无符号整型x是否小于1 => 无符号整型x是否等于0
(Leq(8U|16U|32U|64U) (MOVDconst [1]) x)  => (Neq(8|16|32|64) (MOVDconst [0]) x)
 //1是否小于无符号整型x => 无符号整型x是否不等于0
```

然后在cmd/internal/ssa/gen这个路径下输入
```
go run *.go
```
编译器会在根据所修改的.rules和Ops.go文件自动生成新的文件
这里会产生的新的opGen.go和rewriteARM64.go


### 4.优化结论

上述优化利用了无整型数与0或1的对比的特殊性，根据ARM64独有的指令进行了优化。使得符合以下该情况 0 < x , x <= 0 , x < 1 , 1 <= x 都会得到优化。使得这些情况下做判断的时候，比以往做对比的时候少用了一条指令。虽然优化较小，但是还是增加了效率。
