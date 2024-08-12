#!/bin/bash
set -ex
if [[ ! $1 =~ windows_arm_7 && ! $1 =~ windows_arm64 ]]; then
    upx -9  --lzma "$1"
else
    echo "跳过$1"
fi