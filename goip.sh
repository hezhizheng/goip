#!/bin/bash

set -e

echo "开始下载 IP 数据库到临时目录..."

TEMP_FILE="/tmp/Merged-IP.mmdb"

# 国内服务器建议使用 gh-proxy.com 代理 下载飞快
wget -O "$TEMP_FILE" https://gh-proxy.com/https://github.com/NetworkCats/Merged-IP-Data/releases/latest/download/Merged-IP.mmdb

echo "下载完成，移动到目标目录..."

# 改成自己的程序运行目录
mv -f "$TEMP_FILE" /xxpath/Merged-IP.mmdb

echo "移动完成，重启服务..."

# 改成自己的 supervisorctl 配置名称
supervisorctl restart goip:*

echo "更新完成"