# 从一个布尔极值求解算法看ARM的SIMD优化

**摘要：本文通过一个真实案例（numpy的布尔极值求解算法），描述了使用NEON进行算法优化的过程以及一些关键技巧，相对于使用编译器对C代码做优化，性能提升了大约80%。 本文介绍的内容对需要用到NEON实现高性能计算的开发者很有帮助。**

硬件配置：鲲鹏(ARM64)服务器

1.布尔数组极值的求解问题

对一个一维向量，numpy提供了一个argmax函数用于求解数组中元素最大值对应的索引

```python
import numpy as np

a = np.array([3, 1, 2, 4, 6, 1])

#取出a中元素最大值所对应的索引，此时最大值位6，其对应的位置索引值为4，（索引值默认从0开始）

b=np.argmax(a)

print(b)
```

容易想象，如果向量都是由0和1组成的，那么argmax返回的值就是第一个1出现的索引，这就是接下来要介绍的布尔向量

2.Bool_argmax的C语言标准实现，在numpy中是这样实现的

```c
static int
BOOL_argmax(npy_bool *ip, npy_intp n, npy_intp *max_ind,
            `PyArrayObject *NPY_UNUSED(aip))

{
    npy_intp i = 0;
    for (; i < n; i++) {
        if (ip[i]) {
            *max_ind = i;
            return 0;
        }
    }
    *max_ind = 0;
    return 0;
}
```

算法的基本思想很简单，线性扫描一维数组ip，如果位置对应的值是1，就保存最大值索引至max_ind，对于数据量小的数组，这样做的效率还是蛮高的，但是在大数据处理中，会出现大规模的0值，从而对该函数的求解性能提出了要求。如何进一步优化？在ARM平台运用特有的Neon并行指令集貌似可以提高运算效率(官方说法是4倍)，那就尝试一下吧。

3.Bool_argmax的Neon intrinsics实现

```c
int32_t sign_mask(uint8x16_t input)
{
    const int8_t __attribute__ ((aligned (16))) xr[8] = {-7,-6,-5,-4,-3,-2,-1,0};
    uint8x8_t mask_and = vdup_n_u8(0x80);
    int8x8_t mask_shift = vld1_s8(xr);

    uint8x8_t lo = vget_low_u8(input);
    uint8x8_t hi = vget_high_u8(input);

    lo = vand_u8(lo, mask_and);
    lo = vshl_u8(lo, mask_shift);

    hi = vand_u8(hi, mask_and);
    hi = vshl_u8(hi, mask_shift);

    lo = vpadd_u8(lo,lo);
    lo = vpadd_u8(lo,lo);
    lo = vpadd_u8(lo,lo);

    hi = vpadd_u8(hi,hi);
    hi = vpadd_u8(hi,hi);
    hi = vpadd_u8(hi,hi);

    return ((hi[0] << 8) | (lo[0] & 0xFF));
}

static int
BOOL_argmax(npy_bool *ip, npy_intp n, npy_intp *max_ind,
            PyArrayObject *NPY_UNUSED(aip))

{
    npy_intp i = 0;
    #if defined(__ARM_NEON__) || defined (__ARM_NEON)
        uint8x16_t zero = vdupq_n_u8(0);
        for(; i < n - (n % 32); i+=32) {
            uint8x16_t d1 = vld1q_u8((char *)&ip[i]);
            uint8x16_t d2 = vld1q_u8((char *)&ip[i + 16]);
            d1 = vceqq_u8(d1, zero);
            d2 = vceqq_u8(d2, zero);
            if(sign_mask(vminq_u8(d1, d2)) != 0xFFFF) {
                break;
            }
        }
	#endif
    for (; i < n; i++) {
        if (ip[i]) {
            *max_ind = i;
            return 0;
        }
    }
    *max_ind = 0;
    return 0;
}
```

这段实现有两个关键点:

- SIMD指令最多只能并行处理128bit的数据，也就是16个char(short)型数据，所以数据规模对32余之后的元素还是需要通过常规C语言处理，所以会看到SIMD实现和C实现并存的情况。
- 数据组装的原理，每次从数组中取两组数据，每组数据包含16个char型数据，对这两组数据进行掩码运算，如果其中一组数据包含1，运算结果就不会等于0xFFFF，直接break退出循环。

实现之后还得看看效果，用numpy自带的benchmark测试结果如下：

```C
       before           after         ratio
     [3f11db40]       [00b21d1b]
     <master>         <neon-argmax>
-       161±0.3μs       47.7±0.5μs     0.30  bench_reduce.ArgMax.time_argmax(<class 'bool'>)
```

优化的效果还是蛮明显的，有70%的性能提升，那是不是到这里就结束了？这段代码提交社区后，有熟悉opencv的老司机提出还有进一步的优化空间，不知道大家发现没有，sign_mask函数用了太多指令了！！！其实掩码操作不需要执行那么多的低位和高位累加移位操作，可以简化为如下代码：

```c
int32_t _mm_movemask_epi8_neon(uint8x16_t input)
{
    int8x8_t m0 = vcreate_s8(0x0706050403020100ULL);
    uint8x16_t v0 = vshlq_u8(vshrq_n_u8(input, 7), vcombine_s8(m0, m0));
    uint64x2_t v1 = vpaddlq_u32(vpaddlq_u16(vpaddlq_u8(v0)));
    return (int)vgetq_lane_u64(v1, 0) + ((int)vgetq_lane_u64(v1, 1) << 8);
}
```

优化之后性能又提升了8%，到此为止社区终于接受了本次提交。