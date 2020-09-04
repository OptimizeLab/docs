# 在openEuler上编译OpenCV 4.4.0 
OpenCV（Open Source Computer Vision Library）是跨平台计算机视觉和机器学习软件库，基于 BSD 许可授权发行，可以在商业和研究领域中免费使用。 

OpenCV 用 C++ 和 C 语言编写，具有 C ++、Python、Java、C#、Ruby、GO 和 MATLAB 等接口，并支持 Windows、Linux、Android 和 Mac OS 等操作系统。OpenCV 可用于开发实时的图像处理、计算机视觉以及模式识别程序。 

# OpenCV 4.4.0 

2020年7月18日，OpenCV官网发布了OpenCV的最新版OpenCV4.4.0。以下为新版主要更新内容： 

#### 1、SIFT算法更新 

SIFT（Scale-Invariant Feature Transform，尺度不变特征变换）算法移至主存储库，支持免费使用（SIFT 的专利已过期）。 

#### 2、DNN模块更新 


* 改进的图层/激活/支持更多模型： 

    * 支持最新的 Yolo v4 ： [＃17148](https://github.com/opencv/opencv/issues/17148)
    * ONNX ：支持 Resnet_backbone （Torchvision）[＃16887](https://github.com/opencv/opencv/pull/16887)
    * 支持 EfficientDet 模型 ： [＃17384](https://github.com/opencv/opencv/pull/17384)
* 新示例demo： 
    * 增加文本识别示例： [C++](https://github.com/opencv/opencv/pull/16941)/ [Python](https://github.com/opencv/opencv/pull/16955)
    * 支持 FlowNet2 optical flow： [＃16575](https://github.com/opencv/opencv/pull/16575)
* 英特尔推理引擎后端 ： 
    * 增加了对 OpenVINO 2020.3 LTS / 2020.4 版本的支持 
    * 计划在下一版本中删除对 NN Builder API 的支持 
* 大量针对CUDA的支持和优化 
#### 3、G-API模块 


* 在 OpenCV 后端引入了用于状态内核的新 API ：GAPI_OCV_KERNEL_ST 
* 在 G-API 模块中增加了面向视频的操作 ：goodFeaturesToTrack，buildOpticalFlowPyramid，calcOpicalFlowPyrLK 
* 增加了图像处理内核 ：Laplacian 和双边过滤器 
* 修复了 G-API 的 OpenCL 后端中的潜在崩溃 
#### 4、其他更新 


* Obj-C/Swift 绑定 ： [＃17165](https://github.com/opencv/opencv/pull/17165)
* BIMEF :  生物启发的多重曝光融合框架，用于弱光图像增强 
* 为文本检测添加笔画宽度变换(Stroke Width Transform,SWT) 
* … 

此外，OpenCV 3.4.11 也已发布，并带有一些 bug 修复和改进。详细内容可查看更新说明： [https://github.com/opencv/opencv/wiki/ChangeLog](https://github.com/opencv/opencv/wiki/ChangeLog)

另外，本次版本更新还释放一个重大信号，OpenCV 计划在下一版本中将授权协议由BSD 2 迁移到 Apache 2，这将消除将 OpenCV 用于商业产品时可能面临的专利风险，对开发者更友好！ 

# 在openEular上编译OpenCV 4.4.0 


下面将介绍在 openEular 上编译 OpenCV 4.4.0 的流程，并记录了其中遇到的一些坑及避坑指南。 

首先创建文件夹
```plain
cd /usr/local/src 
mkdir opencv 
cd opencv 
```
从 OpenCV 社区获取 OpenCV-4.4.0 的源码包 
```plain
wget https://github.com/opencv/opencv/archive/4.4.0.tar.gz 
```
解压，进入并创建 build 文件夹 
```plain
tar -zxvf 3.0.0.tar.gz 
cd opencv-3.0.0/ 
mkdir build 
cd build/ 
```
使用  cmake  编译
```plain
cmake .. 
make -j8 
make install 
```
查看安装 OpenCV 所生成的库文件和头文件。 
```plain
ll /usr/local/lib 
```
>total 36M 
>-rw------- 1 root root 563K Jul 22 00:31 libade.a 
>lrwxrwxrwx 1 root root   24 Jul 22 01:57 libopencv_calib3d.so -> libopencv_calib3d.so.4.4 
>lrwxrwxrwx 1 root root   26 Jul 22 01:57 libopencv_calib3d.so.4.4 ->  libopencv_calib3d.so.4.4.0 
>-rwx------ 1 root root 1.9M Jul 22 01:57 libopencv_calib3d.so.4.4.0 
>lrwxrwxrwx 1 root root   21 Jul 22 00:40 libopencv_core.so -> libopencv_core.so.4.4 
>lrwxrwxrwx 1 root root   23 Jul 22 00:40 libopencv_core.so.4.4 -> libopencv_core.so.4.4.0 
>-rwx------ 1 root root 5.2M Jul 22 00:40 libopencv_core.so.4.4.0 
>lrwxrwxrwx 1 root root   20 Jul 22 01:44 libopencv_dnn.so -> libopencv_dnn.so.4.4 
>lrwxrwxrwx 1 root root   22 Jul 22 01:44 libopencv_dnn.so.4.4 -> libopencv_dnn.so.4.4.0 
>-rwx------ 1 root root 5.4M Jul 22 01:44 libopencv_dnn.so.4.4.0 
>lrwxrwxrwx 1 root root   27 Jul 22 01:50 libopencv_features2d.so -> libopencv_features2d.so.4.4 
>lrwxrwxrwx 1 root root   29 Jul 22 01:50 libopencv_features2d.so.4.4 -> libopencv_features2d.so.4.4.0 

# 避坑指南 

在编译期间，本人遇到了一些问题，记录如下： 


1. 无 cmake 或 cmake 版本低 

开始编译时提示缺少 cmake；或者编译失败，查看 cmake 版本过低，比如 cmake-3.5.1 不行，cmake-3.14 之后可以。此时需要重装较新版本的 cmake，过程如下。 

 获取 cmake-3.17.2 源码包 

```plain
cd /usr/local/src 
wget https://cmake.org/files/v3.17/cmake-3.17.2.tar.gz 
```
若从 cmake 官网下载速度较慢，可以在 gitee 码云提供的 src-openEuler 项目中找到适应openEuler 的 cmake 压缩包： [https://gitee.com/src-openeuler/cmake](https://gitee.com/src-openeuler/cmake)
下载地址 [https://gitee.com/src-openeuler/cmake/blob/master/cmake-3.17.2.tar.gz](https://gitee.com/src-openeuler/cmake/blob/master/cmake-3.17.2.tar.gz)

解压并进入安装目录 
```plain
cd /usr/local/src 
tar -zxvf cmake-3.17.2.tar.gz 
cd cmake-3.17.2 
```
安装 cmake 
```plain
./configure 
make 
make install 
```
测试 cmake 是否安装完成 
```plain
cmake -version 
```
返回内容如下所示，表示安装已经完成。 

```
>cmake version 3.17.2 
>CMake suite maintained and supported by Kitware (kitware.com/cmake). 
#### 2.提示缺少OpenSSL 

>cmake 编译时，提示缺少 OpenSSL 
CMake Error: 
>  Could NOT find OpenSSL, try to set the path to OpenSSL root folder in the system variable OPENSSL_ROOT_DIR (missing: OPENSSL_CRYPTO_LIBRARY)  

 

```plain
进入指定目录下载安装 
```
```plain
cd /usr/local/src 
wget https://www.openssl.org/source/openssl-1.1.1f.tar.gz 
```
若从 OpenSSL 官网下载速度较慢，可以在 gitee码云提供的 src-openEuler 项目中找到适应openEuler 的 OpenSSL 压缩包： [https://gitee.com/src-openeuler/openssl](https://gitee.com/src-openeuler/openssl)
下载地址 [https://gitee.com/src-openeuler/openssl/blob/master/openssl-1.1.1f.tar.gz](https://gitee.com/src-openeuler/openssl/blob/master/openssl-1.1.1f.tar.gz)

解压并进入安装目录 

```plain
tar -xvf openssl-1.1.1f.tar.gz 
cd openssl-1.1.1f 
```
安装 OpenSSL 
```plain
./config 
make 
make install 
```
测试 OpenSSL 是否安装完成 

```plain
openssl -version 
```
>OpenSSL 1.1.1f 



 

参考资料： 


1. [https://github.com/opencv/opencv/wiki/ChangeLog](https://github.com/opencv/opencv/wiki/ChangeLog)



 

