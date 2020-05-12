# tool
### 1. GoFuncsAnalysis.sh
- 用法：bash GoFuncsAnalysis.sh [PACKAGE]
- 描述：统计Go指定包（$GOROOT/src/PACKAGE）中的函数到 PACKAGE_analysis.csv 文件，未指定 PACKAGE 时则统计整个 src 包到 go_analysis.csv 文件
- 参数：-h, --help  显示帮助信息，并退出脚本