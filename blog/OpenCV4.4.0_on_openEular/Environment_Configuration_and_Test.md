在完成OpenCV 4.4.0在openEular上的下载和安装步骤，距离正式使用还需要进行环境配置。 

## 配置环境 

接下来介绍配置环境的详细步骤。 


1. 添加库路径（创建opencv.conf文件） 

输入命令 

```plain
vi /etc/ld.so.conf.d/opencv.conf 
```
在文件中添加 
```plain
/usr/local/lib   //或者安装OpenCV时的路径设置 
```
并保存退出 

1. 添加环境变量 

输入命令 

```plain
vi /etc/profile 
```
在结尾添加 
```plain
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH 
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib 
```
更新环境变量 
```plain
source /etc/profile 
```

1. 其他方式设置 

输入命令 

```plain
vi /etc/bash.bashrc 
```
在结尾添加 
```plain
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/local/lib/pkgconfig 
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib 
```
更新环境变量 
```plain
source  /etc/bash.bashrc 
```

1. 更新系统库缓存 

输入命令 

```plain
ldconfig 
```

1. 查看opencv是否安装成功 

输入命令 

```plain
pkg-config --cflags opencv  //头文件路径 
```
>-I/usr/local/include/opencv4 
```plain
pkg-config --libs opencv  //安装库路径 
```
>-L/usr/local/lib -lopencv_shape -lopencv_stitching -lopencv_objdetect -lopencv_superres -lopencv_videostab -lopencv_calib3d -lopencv_features2d -lopencv_highgui -lopencv_videoio -lopencv_imgcodecs -lopencv_video -lopencv_photo -lopencv_ml -lopencv_imgproc -lopencv_flann -lopencv_core 
```plain
pkg-config opencv --modversion 
```
>4.4.0 

若均显示正常，说明安装正确。 

## demo测试 

在OpenCV内部集成了很多测试demo，可以通过执行这些demo看是否完成opencv的配置。 

```plain
命令进入到samples文件夹 
cd opencv-4.4.0/samples 
可以看到有各种语言的samples 

```
>android                directx  opencl               swift                           winrt 
>CMakeLists.example.in  dnn      opengl               tapi                            winrt_universal 
>CMakeLists.txt         gpu      openvx               va_intel                        wp8 
>cpp                    hal      python               _winpack_build_sample.cmd 
>data                   java     samples_utils.cmake  _winpack_run_python_sample.cmd 

进入C++的example_cmake 

```plain
cd /cpp/example_cmake 
```
由于本人所用的主机没有摄像头等，可以稍微修改 example.cpp 这个程序，之后进行编译执行，即可完成demo测试。 

下面为本人使用的彩色图片转换为灰度图片的测试程序。 

```plain
#include <opencv2/imgproc/imgproc.hpp> 
#include <opencv2/core.hpp> 
#include <opencv2/imgcodecs.hpp> 
#include <opencv2/highgui.hpp> 
#include <iostream> 
using namespace cv; 
using std::cout;	 
using std::endl; 
int main() 
{ 
    std::string image_path = samples::findFile("test.jpg"); 
    Mat img = imread(image_path, IMREAD_COLOR); 
    if(img.empty()) 
    { 
        std::cout << "Could not read the image: " << image_path << std::endl; 
        return 1; 
    } 
    Mat gray_img; 
    cvtColor(img, gray_img, COLOR_BGR2GRAY); 
    bool success = imwrite("test_gray.jpg", gray_img); 
    cout << success << endl; 
    return 0; 
} 
```
测试前后图片如下： 

![test](./images/test.jpg) ![test_gray](./images/test_gray.jpg)



