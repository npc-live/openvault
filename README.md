# OpenVault

加密本地密钥管理工具。替代 `.env` / `~/.zshrc` 硬编码，密钥自动注入 shell，无需每次手动 `export`。

```
openvault set OPENAI_API_KEY    # 隐藏输入，加密存储
node -e "console.log(process.env.OPENAI_API_KEY)"  # 自动注入，直接用
```

---

## 安装

### 方式一：从源码编译（需要 Go 1.22+）

```bash
git clone https://github.com/qing/openvault
cd openvault
make build

# 移动到 PATH
sudo mv openvault /usr/local/bin/
```

### 方式二：直接 go install

```bash
go install github.com/qing/openvault@latest
```

---

## 快速开始

### 1. 初始化

```bash
openvault init
```

在 `~/.config/openvault/vault.db` 创建加密数据库，主密钥自动保存到 macOS Keychain，无需记忆任何密码。

### 2. 配置自动注入（只需一次）

```bash
# zsh
echo 'eval "$(openvault shell-init --shell zsh)"' >> ~/.zshrc
source ~/.zshrc

# bash
echo 'eval "$(openvault shell-init --shell bash)"' >> ~/.bashrc
source ~/.bashrc
```

配置后，你的每条命令执行前，OpenVault 自动把所有密钥注入当前 shell 环境。

### 3. 存入密钥

```bash
openvault set OPENAI_API_KEY
# Enter value for OPENAI_API_KEY: （输入不可见）
# Secret "OPENAI_API_KEY" stored.
```

### 4. 直接使用，无需任何前缀

```bash
# 直接运行，密钥已在环境变量中
npm run dev
docker push myimage
python train.py
```

---

## 命令参考

### `openvault init`

初始化 vault。只需运行一次。

```bash
openvault init
```

### `openvault set <KEY>`

存储一个密钥，终端隐藏输入，不写入 shell history。

```bash
openvault set OPENAI_API_KEY
openvault set AWS_SECRET_ACCESS_KEY
openvault set DATABASE_URL
```

### `openvault get <KEY>`

读取并打印密钥值。

```bash
openvault get OPENAI_API_KEY
```

### `openvault list`

列出所有已存储的密钥名（不显示值）。

```bash
openvault list
# OPENAI_API_KEY
# AWS_SECRET_ACCESS_KEY
# DATABASE_URL
```

### `openvault delete <KEY>`

删除一个密钥。

```bash
openvault delete DATABASE_URL
```

### `openvault env`

以 `export KEY=value` 格式输出所有密钥，用于 `eval` 注入。

```bash
eval "$(openvault env)"
```

### `openvault run <command>`

不依赖 shell hook，直接在命令前加 `openvault run` 注入密钥。适合 CI/CD、Docker 等非交互场景。

```bash
openvault run npm run dev
openvault run docker push myimage
openvault run -- python -c "import os; print(os.environ['OPENAI_API_KEY'])"
```

### `openvault shell-init --shell <zsh|bash>`

输出 shell hook 代码，用于写入配置文件。

```bash
openvault shell-init --shell zsh   # 输出 zsh preexec hook
openvault shell-init --shell bash  # 输出 bash PROMPT_COMMAND hook
```

---

## 安全说明

| 安全项 | 实现 |
|---|---|
| 加密存储 | AES-256-GCM，每个密钥独立随机 nonce |
| 主密钥 | 32 字节随机密钥，托管于 macOS Keychain，永不落盘 |
| 隐藏输入 | 内核级禁 echo，不写入 shell history |
| 文件权限 | DB 文件 `0600`，目录 `0700` |
| 内存清理 | 关闭时清零内存中的密钥 |

**验证加密**（DB 中无明文）：

```bash
strings ~/.config/openvault/vault.db | grep YOUR_SECRET
# 无输出
```

**验证 Keychain 条目**：

```bash
security find-generic-password -s openvault
```

---

## CI/CD 使用

在 CI 环境中没有 Keychain，使用 `openvault run` 配合环境变量直接传入密钥，或使用平台自带的 secrets 管理（GitHub Actions Secrets 等）。OpenVault 主要面向本地开发场景。

---

## 数据存储位置

| 平台 | 路径 |
|---|---|
| macOS | `~/.config/openvault/vault.db` |
| Linux | `~/.config/openvault/vault.db`（主密钥存 `~/.config/openvault/.openvault.key`） |
| 自定义 | 设置 `XDG_CONFIG_HOME` 环境变量 |

---

## 构建

```bash
make build    # 编译到当前目录
make install  # 安装到 $GOPATH/bin
make clean    # 删除编译产物
```

依赖：`github.com/spf13/cobra`、`go.etcd.io/bbolt`、`golang.org/x/term`，无 CGo，`CGO_ENABLED=0` 可静态编译。
