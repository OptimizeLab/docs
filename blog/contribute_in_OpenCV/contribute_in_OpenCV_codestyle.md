## OpenCV贡献代码——代码规范篇

当我们已经准备好为社区贡献的代码后，在提交代码前，还需要检查代码规范。OpenCV的代码主要是由C++完成的，针对OpenCV社区的规范进行总结。这一部分可以在OpenCV官网看到具体的[说明](https://github.com/opencv/opencv/wiki/Coding_Style_Guide)，这里列出比较重要的几点：

1.文件结构

```
|── modules
|   ├── modelue_name 模块的名字
|   │   ├── include
|   │   |   ├──opencv2
|   │   |   |   ├──modelue_name 函数声明的文件通常放在【此文件夹中
|   │   |   ├──modelue_name.hpp 通常会引用modelue_name文件夹中文件，src中的文件引用此文件
|   │   ├── samples 算法的实例代码
|   │   ├── src 函数的实现通常放在这里
|   │   ├── test 测试文件
|   │   ├── tutorials 说明文档的文件夹
|   │   ├──CMakeLists.txt 需要写明此模块编译后的名称，及编译此模块的依赖
```

2.明确代码贡献在哪里

在OpenCV社区有很多 模块，在贡献代码之前需要确定把代码贡献在哪里。常见的为OpenCV贡献新功能代码放在[Opencv_contrib](https://github.com/opencv/opencv_contrib)中；如果为代码修复bug，则贡献到代码本身所在的版块。如果还是不明确时，可以提出一个Issue，在社区上和社区的committer讨论，明确代码放在何处。

3、代码的声明和实现分离

不要将代码的声明和实现都放在.hpp文件中。而是，将声明放在hpp文件中，位于opencv/modules/<module_name>/include/opencv2/<module_name>；实现放在.cpp文件中，位于opencv/modules/<module_name>/src。这点看起来理所当然，但是在具体实现时，需要对代码进行很好的设计，例如在代码中若需要实例化对象，此时需要通过静态方法，通过指针来访问此实例。由于我们的代码最后需要编译成动态链接库为他人使用，因此我们的代码需要保证可以被用户调用。对于需要暴露给用户的函数或类接口，使用CV_EXPORT_W这个宏。

例如：

```c++
static std::shared_ptr<XYZ> XYZ::get(IO io) {
    if (xyz_cs.count(io) == 1) {
        return xyz_cs[io];
    }
    std::shared_ptr<XYZ> XYZ_CS(new XYZ(io));
    xyz_cs[io] = XYZ_CS;
    return xyz_cs[io];
}
```



4.增加单元测试

为社区贡献代码时，很显然，别人是不敢轻易运行我们的代码的，这时，需要通过单元测试保证我们的代码是可靠的。Opencv采用GTest框架进行测试。在我们本地编译测试时，会生成bin文化夹，其中的opencv_test_xxx就是生成的测试的可执行文件，直接运行可以查看测试结果。最后的测试文件在提交PR后，会在CI平台上验证结果。

5.代码规范

除了满足基本的代码规范，在OpenCV社区贡献代码还需要满足如下要求：

a. 在文件的头部，需要写明license，Opencv目前的license是Apache2.0。写明下面的3行，表明遵从OpenCV社区的license：

```
// This file is part of OpenCV project.
// It is subject to the license terms in the LICENSE file found in the top-level directory
// of this distribution and at http://opencv.org/license.html.	
```

b. 社区的代码要使用的namespace是cv::，所有的函数和类都要在这个namespace中。

c.类名采用大驼峰命名方式，函数名以小写字母开头。

d. 缩进为4个空格，不要使用制表符。

e.提交的代码中不能以空格结尾。可以在提交代码前，将opencv/.git/hooks/pre-commit.sample钟命名为opencv/.git/hooks/pre-commit，这样在提交代码时会对结尾空白符进行提示。



在本文的最后，看见大家积极参与社区的讨论，遇到问题也可以在社区进行提问，大家共同解决。

