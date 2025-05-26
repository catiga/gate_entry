#!/bin/bash

echo "📦 编译 Win7 32 位版本..."
GOOS=windows GOARCH=386 go build -o proxy_win7_32.exe main.go

echo "📦 编译 Win7 64 位版本..."
GOOS=windows GOARCH=amd64 go build -o proxy_win7_64.exe main.go

echo "✅ 编译完成！"
