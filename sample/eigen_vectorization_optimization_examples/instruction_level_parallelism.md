## 基于指令级并行（instruction-level parallelism）对单精度浮点数指数函数进行优化
指令级并行(instruction-level parallelism)优化是指当多条指令不存在相关关系时，他们在流水线(pipeline)上可以重叠执行从而提高性能，这种指令序列存在的潜在并行性成为指令级并行。

```cpp
/* calculate y = 1 + r + p5*r**2 + p4*r**3 + p3*r**4 + p2*r**5 + p1*r**6 + p0*r**7*/
/*
  Packet r2 = pmul(r, r);
  Packet r3 = pmul(r2, r);
  Packet y = cst_cephes_exp_p0;
  y = pmadd(y, r, cst_cephes_exp_p1);
  y = pmadd(y, r, cst_cephes_exp_p2);
  y = pmadd(y, r, cst_cephes_exp_p3);
  y = pmadd(y, r, cst_cephes_exp_p4);
  y = pmadd(y, r, cst_cephes_exp_p5);
  y = pmadd(y, r2, r);
  y = padd(y, cst_1);
*/
  // Evaluate the polynomial approximant,improved by instruction-level parallelism.
  Packet y, y1, y2;
  y  = pmadd(cst_cephes_exp_p0, r, cst_cephes_exp_p1);
  y1 = pmadd(cst_cephes_exp_p3, r, cst_cephes_exp_p4);
  y2 = padd(r, cst_1);
  y  = pmadd(y, r, cst_cephes_exp_p2);
  y1 = pmadd(y1, r, cst_cephes_exp_p5);
  y  = pmadd(y, r3, y1);
  y  = pmadd(y, r2, y2);
```
Benchmark code:
```cpp
#include<benchmark/benchmark.h>
#include<Eigen/Core>
using namespace Eigen::internal;

float arr[4]{-0.012244f,-0.22222f,0.22222f,0.7855f}; 
float arr_[4];

static void instruction_no_parallel(benchmark::State &state){
  Packet4f r;
  Packet4f result;
  for(auto _ : state){
    r = pload<Packet4f>(arr);
    benchmark::DoNotOptimize(result = pexp(r));    
    pstore(arr_,result);
  }
}
// register the function as benchmark
BENCHMARK(instruction_no_parallel);
// BENCHMARK(instruction_with_parallel);

BENCHMARK_MAIN();
```

Benchmarks:
```
float32x4_t pexp op on Aarch64 with -O0
------------------------------
Benchmark             Time 
------------------------------
before               315 ns 
after                301 ns 

float32x4_t pexp op on Aarch64 with -O3
------------------------------
Benchmark             Time 
------------------------------
before               19.3 ns 
after                17.8 ns 
```