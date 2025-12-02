# GitHub Actions CI/CD 工作流

本项目包含两个 GitHub Actions 工作流，用于自动化测试、构建和发布。

## 工作流说明

### 1. CI 工作流 (`ci.yml`)

**触发条件：**
- 推送到 `main`、`develop` 或 `feature/**` 分支
- 向 `main` 分支创建 Pull Request

**执行的任务：**
1. **测试 (Test)**
   - 在 Ubuntu、macOS 和 Windows 上运行测试
   - 启用竞态检测 (`-race`)
   - 生成覆盖率报告
   - 上传覆盖率到 Codecov

2. **代码检查 (Lint)**
   - 使用 `golangci-lint` 进行代码质量检查
   - 自动检查常见的代码问题

3. **构建检查 (Build Check)**
   - 构建所有平台的二进制文件（作为验证）
   - 上传构建产物供下载（保留 7 天）

### 2. Release 工作流 (`release.yml`)

**触发条件：**
- 推送以 `v` 开头的 tag（例如 `v1.0.0`, `v2.1.3`）

**执行的任务：**
1. 为每个平台（Linux、macOS、Windows）构建发布版本
2. 编译时注入版本信息：
   - Version: git tag
   - Build Time: 当前 UTC 时间
   - Git Commit: 当前 commit SHA
   - Go Version: Go 版本信息
3. 压缩二进制文件（Linux/macOS: tar.gz, Windows: zip）
4. 创建 GitHub Release 并上传所有平台的构建文件

## 使用指南

### 创建新的 Release

#### 方法 1: 使用 Git 命令行

```bash
# 确保代码已提交并推送
git add .
git commit -m "feat: 准备发布 v1.0.0"
git push origin main

# 创建 tag（推荐使用语义化版本）
git tag -a v1.0.0 -m "Release version 1.0.0"

# 推送 tag 到远程仓库
# 这将自动触发 Release 工作流
git push origin v1.0.0
```

#### 方法 2: 使用 GitHub Web 界面

1. 进入 GitHub 仓库的 Releases 页面
2. 点击 "Draft a new release"
3. 创建一个新的 tag（例如 `v1.0.0`）
4. 填写 release 信息
5. 点击 "Publish release"
6. GitHub Actions 将自动构建并上传所有平台的二进制文件

### 查看工作流状态

1. 进入 GitHub 仓库的 "Actions" 标签页
2. 查看工作流的运行状态
3. 点击特定的工作流运行可以查看详细日志

### 下载构建产物

**从 Release 下载：**
1. 进入 GitHub 仓库的 Releases 页面
2. 选择对应的 Release 版本
3. 下载对应平台的文件：
   - `RestreamerMonitor_linux.tar.gz` (Linux)
   - `RestreamerMonitor_darwin.tar.gz` (macOS)
   - `RestreamerMonitor_windows.zip` (Windows)

**从 CI 构建下载：**
1. 进入 GitHub 仓库的 "Actions" 标签页
2. 选择对应的 CI 工作流运行
3. 在页面底部找到 "Artifacts" 部分
4. 下载 `build-artifacts`

## 配置说明

### 修改触发条件

如果你需要修改工作流的触发条件，可以编辑对应的 `.yml` 文件：

**修改 CI 触发分支：**
```yaml
# 在 ci.yml 中修改
push:
  branches: [ main, develop, 'feature/**' ]
pull_request:
  branches: [ main ]
```

**修改 Release 触发条件：**
```yaml
# 在 release.yml 中修改
push:
  tags:
    - 'v*'  # 只匹配以 v 开头的 tag
    - 'release-*'  # 添加更多匹配模式
```

### 支持的 Go 版本

默认使用 Go 1.21。如需修改，更新两个工作流文件中的：

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.21'  # 修改此处
```

### 添加更多构建平台

如需支持更多平台（例如 ARM64），修改 `release.yml` 的 matrix 部分：

```yaml
strategy:
  matrix:
    include:
      - os: ubuntu-latest
        goos: linux
        goarch: amd64
        suffix: ""
      - os: ubuntu-latest  # 添加 ARM64 Linux
        goos: linux
        goarch: arm64
        suffix: ""
      # 添加更多平台...
```

## 故障排除

### 工作流失败

1. **构建失败**：检查代码是否可以正常编译
   ```bash
   # 本地测试构建
   make build-all
   ```

2. **测试失败**：确保所有测试通过
   ```bash
   go test ./...
   ```

3. **Lint 失败**：运行 golangci-lint
   ```bash
   golangci-lint run
   ```

### Release 工作流未触发

确保 tag 名称正确：
- 必须以 `v` 开头（例如 `v1.0.0`）
- tag 必须推送到 GitHub 远程仓库
- 检查 GitHub Actions 是否启用

### 构建产物未上传

检查工作流日志中的 "Build binary" 步骤，确认构建命令执行成功。

## 最佳实践

1. **使用语义化版本号**：遵循 `vMAJOR.MINOR.PATCH` 格式
   - `v1.0.0`：初始稳定版本
   - `v1.1.0`：新增功能（向后兼容）
   - `v1.0.1`：修复 bug（向后兼容）

2. **在创建 Release 前确保 CI 通过**

3. **编写清晰的 Release 说明**，包括：
   - 新功能
   - 修复的 bug
   - 破坏性变更（如果有）

4. **定期更新依赖**：
   ```bash
   go get -u ./...
   go mod tidy
   ```

## 相关文档

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GoReleaser](https://goreleaser.com/) - 更高级的发布工具
- [语义化版本规范](https://semver.org/)
