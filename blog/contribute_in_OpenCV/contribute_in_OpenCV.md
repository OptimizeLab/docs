
# OpenCV 贡献代码后记


### 1. CI

在贡献代码提交 Pull Request 后，社区会自动进行 **CI**（Continuous Integration， 持续集成）。

**OpenCV 社区的自动 CI 工具为 Buildbot**。BuildBot 是一个系统的自动化编译/测试周期的软件。每当代码有改变，服务器要求不同平台上的客户端立即进行代码构建和测试，收集并报告不同平台的构建和测试结果。BuildBot 由 python 完成，仅依赖 python 环境和 Twisted（一个 python 网络框架），可以在很多平台运行。众多知名的项目都是由 BuildBot 来完成 CI 的，比如：Python, Mozilla, Chromium, WebKit等。

当前使用的 Buildbot 信息如下：

* Buildbot: 0.8.12-11-gacff814
* Twisted: 13.2.0
* Jinja: 2.11.1
* Python: 2.7.6 (default, Nov 23 2017, 15:49:48) [GCC 4.8.4]
* Buildmaster platform: linux2


1. 对于直接贡献到 OpenCV 社区主项目代码，可选的 CI 环境有：
  * Linux：Linux x64 / Linux OpenCL / Linux AVX2 / Linux x64 Debug / Linux32
  * Windows：Win64 / Win64 OpenCL / Win32 / Custom Win
  * ARM：ARMv7 / ARMv8	
  * Android：Android armeabi-v7a / Android pack
  * Mac/iOS：Mac	/ iOS / Custom Mac
  * Others：Docs / Custom		


* 其中必选的是：
  * **Linux x64 / Linux OpenCL / Win64 / Win64 OpenCL / Mac / Android armeabi-v7a / Linux x64 Debug / Docs / iOS**


2. 对于贡献到 OpenCV_contrib的代码，可选的 CI 环境有：
  * Linux：Linux x64 / Linux OpenCL / Linux AVX2 / Linux x64 Debug
  * Windows：Win64 / Win64 OpenCL / Win32
  * ARM：ARMv7 / ARMv8	
  * Android：Android armeabi-v7a / Android pack
  * Mac/iOS：Mac	/ iOS
  * Others：Docs / Custom


* 其中必选的为：
  * **Linux x64 / Win64 / Mac / Android armeabi-v7a / Docs / iOS / Win32**


提交代码后，可以在 http://pullrequest.OpenCV.org 中查看自动编译构建的进程和结果。

编译构建过程中出现 errors 会显示红色，出现 warnings 会显示黄色，均需要修改后重新编译构建。

全部显示绿色则表示全部通过，此时才可以进入到代码 review 阶段。

全部通过 CI 后，几天之内就会进行代码 review，由 OpenCV 的 developer 之一进行，会提出问题或者改进意见，需要代码贡献者尽快回复或改进，如果几个月不回复该 PR 就会被拒绝或者关闭。

这个过程的完整流程图如下：

![opencv-pr-flow](./images/opencv-pr-flow.png)


### 2. errors & warnings

下面记录几个 CI 中遇到的 errors & warnings，都是编码或者撰写文档中经常未注意到的问题：

1. "trailing whitespace" & "new blank line at EOF"

   不好的编码习惯容易造成这个 warning ，在代码的结尾或者空白行处添加了多余的空格或空行。尽管在运行时不会造成错误，但是 CI 确实会判定为不通过。解决办法可以通过使用 IDE 中的空格裁剪功能，在完成全部代码后，对代码进行完整的格式化处理。
   
   在此推荐两个代码风格格式化的工具：Eslint 和 Prerrier。


2. "unused parameter"

   在程序中有些函数参数未使用到，在 CI 中也会判定为 warning，不使用的参数及时删除。
   
   
3. "declaration of 'xxx' shadows a field of 'class xxxx'"
   
   类的成员和构造函数的变量命名一致会出现此 warning，需要改成不同的命名。
   

### 3. 代码 review

下面为 OpenCV 的 developer 在代码 review 时提出的部分问题，也记录下来以供参考：

1. "do not use cv::, since this all is declared inside cv namespace"

   如果程序加了 cv 的命名空间，对于 cv::Mat 等声明就不需要加 cv:: 。
   

2. "move all implementations of the methods into .cpp files"

   将所有的声明都需要放到 .cpp 文件中。
   

3. "remove commented off code from public headers"

   在 public 头文件中去掉所有注释的代码。
   

4. "I just noticed that you submitted pull request from your master branch. It's not very good. Please, make a dedicated feature branch, put your code there and submit pull request from this branch"

   提交  pull request 不能直接从 master 分支提交，需要重新创建一个分支，并且该分支的命名需要跟提交的内容相关。
   
   
参考文献：

1. https://github.com/opencv/opencv/wiki/How_to_contribute
