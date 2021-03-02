##  基于SIMD的Eigen/core模块下ARM平台单/双精度浮点数开方优化
当前Eigen社区中仅有部分开发者对于浮点数开方运算在x86 powerPC mips等平台上进行了向量化优化，因此我们对其在ARM平台上进行优化.(解决了issue1933)。  
下面给出单精度浮点数开方运算向量化实现：
```cpp
// 卡马克快速开方倒数
float InvSqrt (float x){
    float xhalf = 0.5f*x;
    int i = *(int*)&x;
    i = 0x5f3759df - (i>>1);
    x = *(float*)&i;
    // 牛顿迭代
    x = x*(1.5f - xhalf*x*x);
    return x;
}
```
```cpp
#if EIGEN_FAST_MATH

/* Functions for sqrt support packet2f/packet4f.*/
// The EIGEN_FAST_MATH version uses the vrsqrte_f32 approximation and one step
// of Newton's method, at a cost of 1-2 bits of precision as opposed to the
// exact solution. It does not handle +inf, or denormalized numbers correctly.
// The main advantage of this approach is not just speed, but also the fact that
// it can be inlined and pipelined with other computations, further reducing its
// effective latency. This is similar to Quake3's fast inverse square root.
// For more details see: http://www.beyond3d.com/content/articles/8/
template<> EIGEN_STRONG_INLINE Packet4f psqrt(const Packet4f& _x){
  Packet4f half = vmulq_n_f32(_x, 0.5f);
  Packet4ui denormal_mask = vandq_u32(vcgeq_f32(_x, vdupq_n_f32(0.0f)),
                                      vcltq_f32(_x, pset1<Packet4f>((std::numeric_limits<float>::min)())));
  // Compute approximate reciprocal sqrt.
  Packet4f x = vrsqrteq_f32(_x);
  // Do a single step of Newton's iteration. 
  //the number 1.5f was set reference to Quake3's fast inverse square root
  x = vmulq_f32(x, psub(pset1<Packet4f>(1.5f), pmul(half, pmul(x, x))));
  // Flush results for denormals to zero.
  return vreinterpretq_f32_u32(vbicq_u32(vreinterpretq_u32_f32(pmul(_x, x)), denormal_mask));
}

template<> EIGEN_STRONG_INLINE Packet2f psqrt(const Packet2f& _x){
  Packet2f half = vmul_n_f32(_x, 0.5f);
  Packet2ui denormal_mask = vand_u32(vcge_f32(_x, vdup_n_f32(0.0f)),
                                     vclt_f32(_x, pset1<Packet2f>((std::numeric_limits<float>::min)())));
  // Compute approximate reciprocal sqrt.
  Packet2f x = vrsqrte_f32(_x);
  // Do a single step of Newton's iteration.
  x = vmul_f32(x, psub(pset1<Packet2f>(1.5f), pmul(half, pmul(x, x))));
  // Flush results for denormals to zero.
  return vreinterpret_f32_u32(vbic_u32(vreinterpret_u32_f32(pmul(_x, x)), denormal_mask));
}

#else 
template<> EIGEN_STRONG_INLINE Packet4f psqrt(const Packet4f& _x){return vsqrtq_f32(_x);}
template<> EIGEN_STRONG_INLINE Packet2f psqrt(const Packet2f& _x){return vsqrt_f32(_x); }
#endif
```

test with google/benchmark:
```cpp
#include<benchmark/benchmark.h>
#include<Eigen/Core>
using namespace Eigen::internal;

float arr[4] = {2.0f,4.0f,6.0f,7.0f};
float arr_[4];

static void sqrt_std(benchmark::State& state){
  for(auto _ : state){
    for(int i=0;i<4;i++){
        arr_[i] = std::sqrt(arr[i]);
    }
  }
}

static void psqrt_simd(benchmark::State& state){
  Packet4f _x;
  for(auto _ : state){
    _x = pload<Packet4f>(arr);  
    _x = psqrt(_x);
    pstore(arr_, _x);
    // if(arr_[0]==0) arr[0] = 2.0;
    // std::cout<<arr_[0];
  }
}

BENCHMARK(sqrt_std);
BENCHMARK(psqrt_simd);
BENCHMARK_MAIN();
```

Benchmarks:
```
Run on (8 X 2600 MHz CPU s)
CPU Caches:
  L1 Data 64 KiB (x8)
  L1 Instruction 64 KiB (x8)
  L2 Unified 512 KiB (x8)
  L3 Unified 32768 KiB (x1)
Load Average: 0.00, 0.06, 0.02

float32x4_t / float32x2_t:
----------------------------
Benchmark            Time            
----------------------------
sqrt_std      23.2/10.0 ns      
psqrt_simd    13.1/6.56 ns 
psqrt_quake        3.50 ns    

float64x2_t:
----------------------------
Benchmark            Time   
----------------------------
sqrt_std           15.8 ns   
psqrt_simd         15.8 ns 
psqrt_quake        6.94 ns     
```