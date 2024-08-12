#!/bin/bash
set -ex
if [[ ! $1 =~ windows_arm_7 && ! $1 =~ windows_arm64 && ! $1 =~ darwin ]]; then
    /tmp/upx-4.2.4-amd64_linux/upx -9  --lzma "$1"
else
    echo "跳过$1"
fi