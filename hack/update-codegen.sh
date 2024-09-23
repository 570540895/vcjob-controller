#!/usr/bin/env bash

# 设置脚本在执行过程中遇到任何错误时立即退出
set -o errexit
# 设置脚本在使用未定义的变量时立即退出
set -o nounset
# 设置脚本在管道中的任何一个命令失败时立即退出
set -o pipefail

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.

$GOPATH/pkg/mod/k8s.io/code-generator@v0.19.2/generate-groups.sh "all" \
  github.com/570540895/vcjob-controller/pkg/client \
  github.com/570540895/vcjob-controller/pkg/apis \
  "batch:v1alpha1" \
  --go-header-file $GOPATH/github.com/570540895/vcjob-controller/hack/boilerplate.go.txt \
  --output-base $GOPATH

# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt
