#!/bin/bash

# 定义要处理的 Go 源文件数组
files=(
  "math.go"
  # 在此添加更多的文件名
)

# 遍历文件数组,并执行可执行程序
for file in "${files[@]}"; do
  echo "Processing file: $file"
  ./instrument "$file"
done