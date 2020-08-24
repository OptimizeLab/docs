# OptimizeLab Overlay 软件仓库用户指南
> 充分发挥硬件性能

OptimizeLab Overlay 软件仓库当前主要针对 aarch64 架构提供关键应用和函数库的高性能预编译版本，采用 Launchpad PPA 的方式提供，以 Ubuntu 18.04 LTS 和 Ubuntu 20.04 LTS 为基础版本。

仓库中的软件包都经过了精心挑选和整理，在使用过程中不会与相应版本系统中现有的的软件产生冲突，并持续跟进维护安全更新和质量更新。在开发的过程中，我们已尽可能避免对开发者造成干扰，力求使用本仓库后，动态编译的应用程序仍能在多数情况下保持和原装系统一致的ABI兼容性。也是因为这样，用户在使用本仓库软件后，可以安全平顺地升级到更高版本的 Ubuntu LTS。

本仓库分为四个组件，随着开发工作的不断深入，我们会根据用户反馈和开发计划增设更多分类。您可以按照需要安装以下任意数量的仓库，但是它们大多数都依赖于 `base` 组件的启用。

### base

顾名思义，`base` 组件是大多数其他组件仓库的基础，建议全部用户使用。如不使用这个仓库，其他仓库应仍能独立使用，但是部分性能优化有赖于本仓库中做出的基础性改进，会导致优化效果不明显甚至无效果。

* 工具链软件(GCC 8 等)，带有 ARMv8.1 LSE 特性支持，不自动替换系统默认编译器。
* Glibc: 使用新工具链编译（GCC 8 vs GCC 7)，并启用增强的优化配置，特定用例下性能提升超 50%
* jemalloc: 提升忙碌情况下 alloc/dealloc 内存时的原子操作性能和稳定性
* valgrind 3.15.0 （新版本）
* 其他系统底层依赖的软件和函数库

要添加这个仓库：
```
sudo add-apt-repository ppa:optimizelab/optimizelab-base
sudo apt update
```

### database

`database` 组件主要包含全新编译的数据库软件包，建议数据库重负载用户选用。

* mysql-8.0 优化编译，较原系统自带软件包大幅提升原子操作性能
* sqlite3 优化编译版本
* 计划推出 postgresql 优化编译版本
* 计划新增 openGauss 数据库软件

要添加这个仓库：
```
sudo add-apt-repository ppa:optimizelab/optimizelab-database
sudo apt update
```

### media

`media` 组件是针对选定的一部分媒体相关软件包进行了优化和升级，主要推荐有编码需求的用户使用，其中的部分软件对音视频播放也有提升作用，追求高性能的用户也可以考虑使用。

* ffmpeg 4.2.2 (大版本更新)，提升编码效率、增加功能，对使用 ffmpeg 解码的软件也有效率提升
* mpv: 基于新版本 ffmpeg 编译，提升 ARM 平台 CPU 解码流畅度
* obs-studio 25.0.3: 新版本，同时使用新版本 ffmpeg
* x264 更新至最新版本
* x265 更新至最新版本
* 其他必要的依赖软件更新

要添加这个仓库：
```
sudo add-apt-repository ppa:optimizelab/optimizelab-media
sudo apt update
```

### science

`science` 组件以科学计算核心软件包为目标，整体瞄准科学计算、HPC 等使用场景，由于其中包含较多的基础计算函数库，对一些计算密集、但并非科学计算应用场景也有明显的性能提升。

* openblas optimized build (also new version)
* openblas 0.3.7: 新版本，优化 ARM 平台编译参数
* blis 0.7.0：新版本
* lapack: 优化 ARM 平台编译参数
* eigen3 3.3.7: 新版本
* julia 1.4.1: 首次引入至 18.04，后续视情况可能迁移至独立组件
* numpy 1.17.4: 新版本，并合入部分上游性能优化补丁，即将随上游引入对 SIMD 优化
* scipy 1.3.3: 新版本
* numba 0.48.0: 新版本
* 其他常用性能相关基础库，如 chafa, fftw3, gmp, psimd, simdjson 等

要添加这个仓库：
```
sudo add-apt-repository ppa:optimizelab/optimizelab-science
sudo apt update
```

## 脚注

这个仓库自身在 Apache-2 许可证下分发，详情见 ``LICENSE`` 文件内容。仓库里的所有软件都保持其原有的许可协议。

如有任何疑问或请求，请向我们提交 issue
