#!/bin/bash

# Remove old binaries (if any)
rm -rf dist

pushd go

env GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o "../dist/vsr_linux_i386"              # Linux i386
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "../dist/vsr_linux_x86_64"          # Linux 64bit
env GOOS=linux GOARCH=arm GOARM=5 go build -ldflags="-s -w" -o "../dist/vsr_linux_arm"       # Linux armv5/armel/arm (it also works on armv6)
env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o "../dist/vsr_linux_armhf"     # Linux armv7/armhf
env GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "../dist/vsr_linux_aarch64"         # Linux armv8/aarch64
env GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -o "../dist/vsr_freebsd_x86_64"      # FreeBSD 64bit
env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "../dist/vsr_darwin_x86_64"        # Darwin 64bit
# env GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o "../dist/vsr_windows_i386.exe"      # Windows 32bit
# env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "../dist/vsr_windows_x86_64.exe"  # Windows 64bit

popd
