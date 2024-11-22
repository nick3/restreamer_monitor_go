.PHONY: build build-all clean

# 项目基本信息
NAME := RestreamerMonitor

# NOTE: This Makefile is used to build the project for different platforms.
# Available targets:
#   - build: Build for current platform
#   - build-all: Build for all supported platforms
#   - clean: Remove all build artifacts

# 系统信息
GOOS_LINUX := linux
GOOS_MAC := darwin
GOOS_WINDOWS := windows
GOARCH := amd64

# 可执行文件后缀（Windows 需要 .exe）
BINARY_SUFFIX_WINDOWS := .exe
BINARY_SUFFIX_LINUX :=
BINARY_SUFFIX_MAC :=

# 获取版本信息 (简化方式)
VERSION := $(shell git describe --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# 编译参数
LDFLAGS := -X 'main.Version=${VERSION}' \
		   -X 'main.BuildTime=${BUILD_TIME}' \
		   -X 'main.GitCommit=${GIT_COMMIT}' \
		   -X 'main.GoVersion=${GO_VERSION}' \
		   -w -s

# 默认编译当前平台
build:
	@echo "Building for current platform..."
	go build -trimpath -ldflags "${LDFLAGS}" -o bin/$(NAME) ./main/main.go
	@echo "Done."

# 编译所有平台
build-all:
	@echo "Building for all platforms..."
	# Linux
	GOOS=${GOOS_LINUX} GOARCH=${GOARCH} go build -trimpath -ldflags "${LDFLAGS}" \
		-o bin/$(NAME)_${GOOS_LINUX}${BINARY_SUFFIX_LINUX} ./main/main.go
	# MacOS
	GOOS=${GOOS_MAC} GOARCH=${GOARCH} go build -trimpath -ldflags "${LDFLAGS}" \
		-o bin/$(NAME)_${GOOS_MAC}${BINARY_SUFFIX_MAC} ./main/main.go
	# Windows
	GOOS=${GOOS_WINDOWS} GOARCH=${GOARCH} go build -trimpath -ldflags "${LDFLAGS}" \
		-o bin/$(NAME)_${GOOS_WINDOWS}${BINARY_SUFFIX_WINDOWS} ./main/main.go
	# Windows GUI模式 (无控制台窗口)
	# GOOS=${GOOS_WINDOWS} GOARCH=${GOARCH} go build -trimpath -ldflags "-H windowsgui ${LDFLAGS}" \
	# 	-o bin/w$(NAME)_${GOOS_WINDOWS}${BINARY_SUFFIX_WINDOWS} ./main/main.go
	@echo "Done."

# 清理编译产物
clean:
	@echo "Cleaning build artifacts..."
	go clean -v
	rm -rf bin/
	@echo "Done."