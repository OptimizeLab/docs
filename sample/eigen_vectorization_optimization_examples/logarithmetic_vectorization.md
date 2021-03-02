## 基于SIMD的Eigen/core模块下多平台双精度浮点数对数函数优化
指数函数和对数函数在神经网络的训练中起到重要的作用，常在神经网络的激活函数和损失函数中使用，如sigmod，softmax和交叉熵损失等。完成了 NEON、SSE、AVX、AVX512模块下双精度浮点数对数函数的向量化实现。

```cpp
/* Natural logarithm
 * Computes log(x) as log(2^e * m) = C*e + log(m), where the constant C =log(2)
 * and m is in the range [sqrt(1/2),sqrt(2)). In this range, the logarithm can
 * be easily approximated by a polynomial centered on m=1 for stability.
 * Returns the base e (2.718...) logarithm of m.
 * The argument is separated into its exponent and fractional
 * parts.  If the exponent is between -1 and +1, the logarithm
 * of the fraction is approximated by
 *
 *     log(1+x) = x - 0.5 x**2 + x**3 P(x)/Q(x).
 *
 * Otherwise, setting  z = 2(x-1)/x+1),
 *                     log(x) = z + z**3 P(z)/Q(z).
 * 
 * for more detail see: http://www.netlib.org/cephes/
 */
template <typename Packet>
EIGEN_DEFINE_FUNCTION_ALLOWING_MULTIPLE_DEFINITIONS
EIGEN_UNUSED
Packet plog_double(const Packet _x)
{
  Packet x = _x;

  const Packet cst_1              = pset1<Packet>(1.0);
  const Packet cst_half           = pset1<Packet>(0.5);
  // The smallest non denormalized float number.
  const Packet cst_min_norm_pos   = pset1frombits<Packet>( static_cast<uint64_t>(0x0010000000000000ull));
  const Packet cst_minus_inf      = pset1frombits<Packet>( static_cast<uint64_t>(0xfff0000000000000ull));
  const Packet cst_pos_inf        = pset1frombits<Packet>( static_cast<uint64_t>(0x7ff0000000000000ull));

 // Polynomial Coefficients for log(1+x) = x - x**2/2 + x**3 P(x)/Q(x)
 //                             1/sqrt(2) <= x < sqrt(2)
  const Packet cst_cephes_SQRTHF = pset1<Packet>(0.70710678118654752440E0);
  const Packet cst_cephes_log_p0 = pset1<Packet>(1.01875663804580931796E-4);
  const Packet cst_cephes_log_p1 = pset1<Packet>(4.97494994976747001425E-1);
  const Packet cst_cephes_log_p2 = pset1<Packet>(4.70579119878881725854E0);
  const Packet cst_cephes_log_p3 = pset1<Packet>(1.44989225341610930846E1);
  const Packet cst_cephes_log_p4 = pset1<Packet>(1.79368678507819816313E1);
  const Packet cst_cephes_log_p5 = pset1<Packet>(7.70838733755885391666E0);

  const Packet cst_cephes_log_r0 = pset1<Packet>(1.0);
  const Packet cst_cephes_log_r1 = pset1<Packet>(1.12873587189167450590E1);
  const Packet cst_cephes_log_r2 = pset1<Packet>(4.52279145837532221105E1);
  const Packet cst_cephes_log_r3 = pset1<Packet>(8.29875266912776603211E1);
  const Packet cst_cephes_log_r4 = pset1<Packet>(7.11544750618563894466E1);
  const Packet cst_cephes_log_r5 = pset1<Packet>(2.31251620126765340583E1);

  const Packet cst_cephes_log_q1 = pset1<Packet>(-2.121944400546905827679e-4);
  const Packet cst_cephes_log_q2 = pset1<Packet>(0.693359375);

  // Truncate input values to the minimum positive normal.
  x = pmax(x, cst_min_norm_pos);

  Packet e;
  // extract significant in the range [0.5,1) and exponent
  x = pfrexp(x,e);
  
  // Shift the inputs from the range [0.5,1) to [sqrt(1/2),sqrt(2))
  // and shift by -1. The values are then centered around 0, which improves
  // the stability of the polynomial evaluation.
  //   if( x < SQRTHF ) {
  //     e -= 1;
  //     x = x + x - 1.0;
  //   } else { x = x - 1.0; }
  Packet mask = pcmp_lt(x, cst_cephes_SQRTHF);
  Packet tmp = pand(x, mask);
  x = psub(x, cst_1);
  e = psub(e, pand(cst_1, mask));
  x = padd(x, tmp);

  Packet x2 = pmul(x, x);
  Packet x3 = pmul(x2, x);

  // Evaluate the polynomial approximant , probably to improve instruction-level parallelism.
  // y = x * ( z * polevl( x, P, 5 ) / p1evl( x, Q, 5 ) );
  Packet y, y1, y2,y_;
  y  = pmadd(cst_cephes_log_p0, x, cst_cephes_log_p1);
  y1 = pmadd(cst_cephes_log_p3, x, cst_cephes_log_p4);
  y  = pmadd(y, x, cst_cephes_log_p2);
  y1 = pmadd(y1, x, cst_cephes_log_p5);
  y_ = pmadd(y, x3, y1);

  y  = pmadd(cst_cephes_log_r0, x, cst_cephes_log_r1);
  y1 = pmadd(cst_cephes_log_r3, x, cst_cephes_log_r4);
  y  = pmadd(y, x, cst_cephes_log_r2);
  y1 = pmadd(y1, x, cst_cephes_log_r5);
  y  = pmadd(y, x3, y1);

  y_ = pmul(y_, x3);
  y  = pdiv(y_, y);

  // Add the logarithm of the exponent back to the result of the interpolation.
  y1  = pmul(e, cst_cephes_log_q1);
  tmp = pmul(x2, cst_half);
  y   = padd(y, y1);
  x   = psub(x, tmp);
  y2  = pmul(e, cst_cephes_log_q2);
  x   = padd(x, y);
  x   = padd(x, y2);

  Packet invalid_mask = pcmp_lt_or_nan(_x, pzero(_x));
  Packet iszero_mask  = pcmp_eq(_x,pzero(_x));
  Packet pos_inf_mask = pcmp_eq(_x,cst_pos_inf);
  // Filter out invalid inputs, i.e.:
  //  - negative arg will be NAN
  //  - 0 will be -INF
  //  - +INF will be +INF
  return pselect(iszero_mask, cst_minus_inf,
                              por(pselect(pos_inf_mask,cst_pos_inf,x), invalid_mask));
}
```
实现pfrexp函数如下：
```cpp
template<typename Packet> EIGEN_STRONG_INLINE Packet
pfrexp_double(const Packet& a, Packet& exponent) {
  typedef typename unpacket_traits<Packet>::integer_packet PacketI;
  const Packet cst_1022d = pset1<Packet>(1022.0);
  const Packet cst_half = pset1<Packet>(0.5);
  const Packet cst_inv_mant_mask  = pset1frombits<Packet>(~0x7ff0000000000000u);
  exponent = psub(pcast<PacketI,Packet>(plogical_shift_right<52>(preinterpret<PacketI>(a))), cst_1022d);
  return por(pand(a, cst_inv_mant_mask), cst_half);
}
```
Benchmarks:  
AArch64 with NEON:
```
Run on (8 X 2600 MHz CPU s)
CPU Caches:
  L1 Data 64 KiB (x8)
  L1 Instruction 64 KiB (x8)
  L2 Unified 512 KiB (x8)
  L3 Unified 32768 KiB (x1)
Load Average: 0.12, 0.17, 0.13
---------------------------
Benchmark           Time 
---------------------------
log_std           55.8 ns 
plog_simd         26.0 ns 
```

x86_64 with SSE/AVX/AVX512:
```
Run on (8 X 3000 MHz CPU s)
CPU Caches:
  L1 Data 32 KiB (x4)
  L1 Instruction 32 KiB (x4)
  L2 Unified 1024 KiB (x4)
  L3 Unified 30976 KiB (x1)
Load Average: 0.05, 0.01, 0.00
----------------------------------
Benchmark               Time 
----------------------------------
log_std         31.7/65.6/131.0 ns 
plog_simd       11.1/12.0/12.6  ns 
```