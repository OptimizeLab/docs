#!/bin/bash
case "$1" in
    -h|--help|?)
    echo "Usage: $0 [PACKAGE]"
    echo "Count all funcs in the specified package of Go"
    echo ""
    echo "With no PACKAGE, Count all funcs in the root dir of Go."
    echo ""
    echo "  -h, --help     display this help and exit"
    exit 0
;;
esac

pkg=$1
if [ "${GOROOT}" == "" ] || [ ! -d  "${GOROOT}/src" ] ; then
	echo "Please set the correct 'GOROOT' first !!"
	exit 1
fi
src_dir="${GOROOT}/src"
pkg_dir="${src_dir}"
file_name="go_analysis.csv"
if [ "${pkg}" != "" ]; then
	pkg_dir="${src_dir}/${pkg}"
	if [ ! -d  "${pkg_dir}" ] ; then
		echo "Package dir `${pkg_dir}` not found !"
		exit 1
	fi

	file_name="${pkg}_analysis.csv"
fi

echo "编号,包名,子包,函数名,依赖关系（关键调用）,ARM与X86优化差异分析,涉及的数学理论,SIMD优化,算法改进（时间/空间）,cache优化,位宽对齐,SSA规则优化,汇编优化,CleanCode,备注" > ${file_name}
num=0
IFS=$'\n'
for fl in `find ${src_dir}/${pkg} -name "*.go" | grep -v "_test.go"`; do
	full_name=${fl#*${src_dir}/}
	pkg_name=${full_name%%/*}
	suf_sub=${full_name%/*}
	sub_pkg=${suf_sub#*/}

	if [ "${sub_pkg}" == "${pkg_name}" ];then
		sub_pkg=""
	fi
	for line in `grep "^func " ${fl}`; do
		num=`expr ${num} + 1`
		full_func=${line#*func }
		func_name=${full_func%%(*}
		if [ "${func_name}" == "" ];then
			pre_func=${full_func#*) }
			func_name=${pre_func%%(*}
		fi
		echo "${num},${pkg_name},${sub_pkg},${func_name}" >> ${file_name}
	done
done
