# 浮点数与0比较的SSA规则优化
### 安装包和源码准备
- [Golang发行版 1.12.1 ARM64版](https://golang.org/dl/)安装
- [Golang源码仓库](https://go.googlesource.com/go)下载
```bash
$ git clone https://go.googlesource.com/go
$ cd go/src
```
- 硬件配置：鲲鹏(ARM64)服务器
### 1. 浮点数与0比较性能问题
在代码开发中，经常会出现将变量与0比较的场景，比如为用户返回剩余金额大于0的列表，因此可能会用到如下函数：
```go
func comp(x float64, arr []int) {
    for i := 0; i < len(arr); i++ {
        if x > 0 {
            arr[i] = 1
        }
    }
}
```


使用如下compile命令查看该函数的汇编代码：
```bash
go tool compile -S main.go
```
```assembly
"".comp STEXT size=80 args=0x20 locals=0x0 leaf
        0x0000 00000 (main.go:3)        TEXT    "".comp(SB), LEAF|NOFRAME|ABIInternal, $0-32
        0x0000 00000 (main.go:3)        FUNCDATA        ZR, gclocals·09cf9819fc716118c209c2d2155a3632(SB)
        0x0000 00000 (main.go:3)        FUNCDATA        $1, gclocals·69c1753bd5f81501d95132d08af04464(SB)
        0x0000 00000 (main.go:3)        FUNCDATA        $3, gclocals·568470801006e5c0dc3947ea998fe279(SB)
        0x0000 00000 (main.go:4)        PCDATA  $2, ZR
        0x0000 00000 (main.go:4)        PCDATA  ZR, ZR
        0x0000 00000 (main.go:4)        MOVD    "".arr+16(FP), R0
        0x0004 00004 (main.go:4)        PCDATA  $2, $1
        0x0004 00004 (main.go:4)        PCDATA  ZR, $1
        0x0004 00004 (main.go:4)        MOVD    "".arr+8(FP), R1          // 取数组地址
        0x0008 00008 (main.go:4)        FMOVD   "".x(FP), F0              // 将参数x放入F0寄存器
        0x000c 00012 (main.go:4)        MOVD    ZR, R2                    // R2 清零
        0x0010 00016 (main.go:4)        JMP     24                        // 第一轮循环直接跳到条件比较 不增加i
        0x0014 00020 (main.go:4)        ADD     $1, R2, R2                // i++
        0x0018 00024 (main.go:4)        CMP     R0, R2                    // i < len(arr) 比较
        0x001c 00028 (main.go:4)        BGE     68                        // i == len(arr) 跳转到末尾
        0x0020 00032 (main.go:5)        FMOVD   ZR, F1                    // 将0复制到浮点寄存器F1
        0x0024 00036 (main.go:5)        FCMPD   F1, F0                    // 将浮点寄存器F0和F1中的值进行比较
        0x0028 00040 (main.go:5)        CSET    GT, R3                    // F0 > F1 -> R3 = 1
        0x002c 00044 (main.go:5)        CBZ     R3, 60                    // R3 == 1 即 x <= 0 跳转到60
        0x0030 00048 (main.go:6)        MOVD    $1, R3                    // x > 0
        0x0034 00052 (main.go:6)        MOVD    R3, (R1)(R2<<3)           // 将切片中值赋值为1
        0x0038 00056 (main.go:6)        JMP     20                        // 跳转到20 即循环操作i++处
        0x003c 00060 (main.go:6)        MOVD    $1, R3                    // x <= 0
        0x0040 00064 (main.go:5)        JMP     20
        0x0044 00068 (<unknown line number>)    PCDATA  $2, $-2
        0x0044 00068 (<unknown line number>)    PCDATA  ZR, $-2
        0x0044 00068 (<unknown line number>)    RET     (R30)
        0x0000 e0 0f 40 f9 e1 0b 40 f9 e0 07 40 fd 02 00 80 d2  ..@...@...@.....
        0x0010 02 00 00 14 42 04 00 91 5f 00 00 eb 4a 01 00 54  ....B..._...J..T
        0x0020 e1 03 67 9e 00 20 61 1e e3 d7 9f 9a 83 00 00 b4  ..g.. a.........
        0x0030 e3 03 40 b2 23 78 22 f8 f7 ff ff 17 e3 03 40 b2  ..@.#x".......@.
        0x0040 f5 ff ff 17 c0 03 5f d6 00 00 00 00 00 00 00 00  ......_.........
        rel 68+0 t=11 +0
```
可以看到对于浮点数与0的比较，上述代码首先将0放入F1寄存器，之后使用FCMPD命令将F0寄存器中的值x与F1寄存器中的0值进行比较

对于长度为100的切片性能如下：
```bash
goos: linux
goarch: arm64
BenchmarkFloatCompare-8         100000000               13.1 ns/op
```

对于浮点数比较的ARM指令[FCMP](http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.dui0068b/Bcfejdgg.html)指令有两种用法：
1. 将两个浮点寄存器中的值进行比较；
2. 将一个浮点寄存器中的值与数值0比较；

可以看到对于FCMP指令，浮点数与0比较是一个特例，不需要将0放入一个浮点寄存器中，可以直接使用FCMP F0, $(0) 进行比较，因此上述生成的汇编代码并不是最优的

### 2. 浮点数与0比较的SSA规则优化
#### 2.1 问题分析
通过Golang的SSA工具进一步分析上述代码：
```bash
GOSSAFUNC=comp go tool compile main.go
```
可以查看Golang生成汇编代码的优化过程，这个过程使用到了[静态单赋值SSA](https://en.wikipedia.org/wiki/Static_single_assignment_form)：
- 更多关于[Golang SSA](https://github.com/golang/go/blob/master/src/cmd/compile/internal/ssa/README.md)

上述命令会生成一个ssa.html，使用浏览器打开：

![image](images/ssa_before_opt.png)

- 该文件展示SSA优化的过程，最后一步是SSA规则优化后的最终形式：

![image](images/ssa_before_opt_result.png)

在图中红色下划线标注处可以看到与上节含义一致的汇编代码。

#### 2.2 FCMP的SSA规则优化
经过对比分析，发现最新的Golang版本已经对上述SSA规则进行了优化，具体优化CL见：
[cmd/compile: optimize arm64 comparison of x and 0.0 with "FCMP $(0.0), Fn"](https://go-review.googlesource.com/c/go/+/164719)
该优化通过SSA规则转换，所有浮点数与0的比较都会受益
#### 2.2 SSA规则优化前后对比
使用最新版的go编译器查看SSA
```bash
GOROOT=/usr/local/src/ssa_end/go; GOSSAFUNC=comp go tool compile main.go
```
![image](images/ssa_after_opt_result.png)
可以看到两条指令变成了一条
- 使用Golang源码生成编译工具请参考案例[Golang在ARM64开发环境配置](../del-env-pre/del-env-pre.md)

#### 2.3 SSA优化规则解析
src/cmd/compile/internal/arm64/ssa.go
```go
case ssa.OpARM64FCMPS0,                     // FCMPS0 -> FCMPS $(0.0), F0
     ssa.OpARM64FCMPD0:                     // FCMPD0 -> FCMPD $(0.0), F0
     p := s.Prog(v.Op.Asm())                // FCMPS | FCMPD
     p.From.Type = obj.TYPE_FCONST          // $(0.0) 的类型为常数
     p.From.Val = math.Float64frombits(0)   // 比较的数 $(0.0)
     p.Reg = v.Args[0].Reg()                // 第二个源操作数，即用于比较的浮点数寄存器F0
```
src/cmd/compile/internal/ssa/gen/ARM64.rules
```bash
// Optimize comparision between a floating-point value and 0.0 with "FCMP $(0.0), Fn"
(FCMPS x (FMOVSconst [0])) -> (FCMPS0 x)                // 浮点数比较操作转换：x(float32)与常数0比较 -> FCMPS0 x
(FCMPS (FMOVSconst [0]) x) -> (InvertFlags (FCMPS0 x))  // 浮点数比较操作转换：常数0与x(float32)比较 -> FCMPS0 x 取反
(FCMPD x (FMOVDconst [0])) -> (FCMPD0 x)                // 浮点数比较操作转换：x(float64)与常数0比较 -> FCMPD0 x
(FCMPD (FMOVDconst [0]) x) -> (InvertFlags (FCMPD0 x))  // 浮点数比较操作转换：常数0与x(float64)比较 -> FCMPD0 x 取反

(LessThanF (InvertFlags x)) -> (GreaterThanF x)         // 浮点数比较条件判断转换：小于取反 -> 大于
(LessEqualF (InvertFlags x)) -> (GreaterEqualF x)       // 浮点数比较条件判断转换：小于等于取反 -> 大于等于
(GreaterThanF (InvertFlags x)) -> (LessThanF x)         // 浮点数比较条件判断转换：大于取反 -> 小于
(GreaterEqualF (InvertFlags x)) -> (LessEqualF x)       // 浮点数比较条件判断转换：大于等于取反 -> 小于等于
```
src/cmd/compile/internal/ssa/gen/ARM64Ops.go
```go
fp1flags  = regInfo{inputs: []regMask{fp}}              // 定义一个寄存器的输入参数mask，此处fp表示所有浮点数寄存器

// 定义操作FCMPS0，将浮点寄存器中的参数(float32)与0进行比较,使用汇编指令FCMPS
{name: "FCMPS0", argLength: 1, reg: fp1flags, asm: "FCMPS", typ: "Flags"},   // arg0 compare to 0, float32
// 定义操作FCMPD0，将浮点寄存器中的参数(float64)与0进行比较
{name: "FCMPD0", argLength: 1, reg: fp1flags, asm: "FCMPD", typ: "Flags"},   // arg0 compare to 0, float64
```
在src/cmd/compile/internal/ssa/gen目录下执行命令：
```bash
 go run *.go
```
得到根据上述规则文件自动生成的opGen.go 和 rewriteARM64.go
#### 2.4 工具自动生成的代码解析
根据ARM64Ops.go生成opGen.op：
```go
//OpARM64FCMPS0
{
    name:   "FCMPS0",                 // 操作名
    argLen: 1,                        // 参数个数
    asm:    arm64.AFCMPS,             // 对应的机器指令
    reg: regInfo{
        inputs: []inputInfo{          // 支持的输入参数寄存器
            {0, 9223372034707292160}, // F0 F1 F2 F3 F4 F5 F6 F7 F8 F9 F10 F11 F12 F13 F14 F15 F16 F17 F18 F19 F20 F21 F22 F23 F24 F25 F26 F27 F28 F29 F30 F31
        },
    },
},
//OpARM64FCMPD0
{
    name:   "FCMPD0",                 // 操作名
    argLen: 1,                        // 参数个数
    asm:    arm64.AFCMPD,             // 对应的机器指令
    reg: regInfo{
        inputs: []inputInfo{          // 支持的输入参数寄存器
            {0, 9223372034707292160}, // F0 F1 F2 F3 F4 F5 F6 F7 F8 F9 F10 F11 F12 F13 F14 F15 F16 F17 F18 F19 F20 F21 F22 F23 F24 F25 F26 F27 F28 F29 F30 F31
        },
    },
},
```

根据ARM64.rules生成rewriteARM64.go：
```bash
// 以下规则会按条挨个匹配，匹配后执行转换
case OpARM64FCMPD:
    return rewriteValueARM64_OpARM64FCMPD_0(v)
case OpARM64FCMPS:
    return rewriteValueARM64_OpARM64FCMPS_0(v)
case OpARM64GreaterEqualF:
    return rewriteValueARM64_OpARM64GreaterEqualF_0(v)
case OpARM64GreaterThanF:
    return rewriteValueARM64_OpARM64GreaterThanF_0(v)
case OpARM64LessEqualF:
    return rewriteValueARM64_OpARM64LessEqualF_0(v)
case OpARM64LessThanF:
    return rewriteValueARM64_OpARM64LessThanF_0(v)

// x(float64)与0比较 转为 FCMPD0 x
func rewriteValueARM64_OpARM64FCMPD_0(v *Value) bool {
    b := v.Block
    _ = b
    // match: (FCMPD x (FMOVDconst [0]))
    // cond:
    // result: (FCMPD0 x)
    for {
        _ = v.Args[1]
        x := v.Args[0]
        v_1 := v.Args[1]
        if v_1.Op != OpARM64FMOVDconst {
            break
        }
        if v_1.AuxInt != 0 {                                      // 如果不是与0比较，则退出
            break
        }
        v.reset(OpARM64FCMPD0)                                    // 修改OpARM64FCMPD指令为OpARM64FCMPD0
        v.AddArg(x)
        return true
    }
    // match: (FCMPD (FMOVDconst [0]) x)
    // cond:
    // result: (InvertFlags (FCMPD0 x))
    for {
        _ = v.Args[1]
        v_0 := v.Args[0]
        if v_0.Op != OpARM64FMOVDconst {
            break
        }
        if v_0.AuxInt != 0 {                                      // 如果不是与0比较，则退出
            break
        }
        x := v.Args[1]
        v.reset(OpARM64InvertFlags)                               // 修改OpARM64FCMPD指令为OpARM64InvertFlags
        v0 := b.NewValue0(v.Pos, OpARM64FCMPD0, types.TypeFlags)  // 添加一个表示OpARM64FCMPD0指令的value(SSA表示一个值)
        v0.AddArg(x)
        v.AddArg(v0)
        return true
    }
    return false
}

// x(float32)与0比较 转为 FCMPS0 x
func rewriteValueARM64_OpARM64FCMPS_0(v *Value) bool {
    b := v.Block
    _ = b
    // match: (FCMPS x (FMOVSconst [0]))
    // cond:
    // result: (FCMPS0 x)
    for {
        _ = v.Args[1]
        x := v.Args[0]
        v_1 := v.Args[1]
        if v_1.Op != OpARM64FMOVSconst {
            break
        }
        if v_1.AuxInt != 0 {              // 如果操作数不为0，退出
            break
        }
        v.reset(OpARM64FCMPS0)            // 修改OpARM64FCMPS指令为OpARM64FCMPS0
        v.AddArg(x)
        return true
    }
    // match: (FCMPS (FMOVSconst [0]) x)
    // cond:
    // result: (InvertFlags (FCMPS0 x))
    for {
        _ = v.Args[1]
        v_0 := v.Args[0]
        if v_0.Op != OpARM64FMOVSconst {
            break
        }
        if v_0.AuxInt != 0 {
            break
        }
        x := v.Args[1]
        v.reset(OpARM64InvertFlags)
        v0 := b.NewValue0(v.Pos, OpARM64FCMPS0, types.TypeFlags)
        v0.AddArg(x)
        v.AddArg(v0)
        return true
    }
    return false
}

// 带反转标志的浮点数比较：invert(x >= 0) 转为 x <= 0
func rewriteValueARM64_OpARM64GreaterEqualF_0(v *Value) bool {
    // match: (GreaterEqualF (InvertFlags x))
    // cond:
    // result: (LessEqualF x)
    for {
        v_0 := v.Args[0]
        if v_0.Op != OpARM64InvertFlags { // 不需反转，此转换规则不适用，退出继续后面的规则
            break
        }
        x := v_0.Args[0]
        v.reset(OpARM64LessEqualF)        // 修改OpARM64GreaterEqualF指令为OpARM64LessEqualF
        v.AddArg(x)
        return true
    }
    return false
}

// 带反转标志的浮点数比较操作：x > 0 转换 x < 0
func rewriteValueARM64_OpARM64GreaterThanF_0(v *Value) bool {
    // match: (GreaterThanF (InvertFlags x))
    // cond:
    // result: (LessThanF x)
    for {
        v_0 := v.Args[0]
        if v_0.Op != OpARM64InvertFlags { // 不需反转，此转换规则不适用，退出继续后面的规则
            break
        }
        x := v_0.Args[0]
        v.reset(OpARM64LessThanF)         // 修改OpARM64GreaterThanF指令为OpARM64LessThanF
        v.AddArg(x)
        return true
    }
    return false
}

// 带反转标志的浮点数比较操作：invert(x <= 0) 转为 x >= 0
func rewriteValueARM64_OpARM64LessEqualF_0(v *Value) bool {
    // match: (LessEqualF (InvertFlags x))
    // cond:
    // result: (GreaterEqualF x)
    for {
        v_0 := v.Args[0]
        if v_0.Op != OpARM64InvertFlags { // 不需反转，此转换规则不适用，退出继续后面的规则
            break
        }
        x := v_0.Args[0]
        v.reset(OpARM64GreaterEqualF)     // 修改OpARM64LessEqualF指令为OpARM64GreaterEqualF
        v.AddArg(x)
        return true
    }
    return false
}

// 带反转标志的浮点数比较操作：invert(x < 0) 转为 x > 0
func rewriteValueARM64_OpARM64LessThanF_0(v *Value) bool {
    // match: (LessThanF (InvertFlags x))
    // cond:
    // result: (GreaterThanF x)
    for {
        v_0 := v.Args[0]
        if v_0.Op != OpARM64InvertFlags { // 不需反转，此转换规则不适用，退出继续后面的规则
            break
        }
        x := v_0.Args[0]
        v.reset(OpARM64GreaterThanF)      // 修改OpARM64LessThanF指令为OpARM64GreaterThanF
        v.AddArg(x)
        return true
    }
    return false
}
```

#### 2.5 结果对比
对于长度为100的切片耗时下降了6.11%：
```bash
name            old time/op  new time/op  delta
FloatCompare-8  13.1ns ± 0%  12.3ns ± 0%  -6.11%  (p=0.008 n=5+5)
```
