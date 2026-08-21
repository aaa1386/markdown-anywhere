# Markown Anywhere 开发方案

## 1. 项目目标

**Markown Anywhere** 用于让 Vault 外的 Markdown 文件也能像普通桌面文件一样直接使用 Obsidian 打开和编辑。

核心流程：

```text
Windows 双击 .md
→ MarkownAnywhere.exe
→ 读取默认 Vault
→ Vault 内：直接打开
→ Vault 外：创建/复用 symlink
→ obsidian://open
→ Obsidian
```

原则：

> **Go EXE 负责核心执行，Obsidian Plugin 负责配置与管理。**

不复制 Markdown，不维护文件副本。

---

## 2. Go Companion

程序：

```text
MarkownAnywhere.exe
```

使用 Go 开发，负责：

- 接收 `.md / .markdown` 文件路径
    
- 读取默认 Vault
    
- Vault 内文件直接打开
    
- Vault 外创建/复用 symlink
    
- 处理同名文件、Unicode/中文路径
    
- 调用 `obsidian://open`
    
- 检测和清理失效 symlink
    
- 注册/取消 Windows Markdown 文件关联
    
- 提供状态查询
    

CLI：

```text
MarkownAnywhere.exe <file>
MarkownAnywhere.exe --register
MarkownAnywhere.exe --unregister
MarkownAnywhere.exe --status --json
MarkownAnywhere.exe --links --json
MarkownAnywhere.exe --cleanup
MarkownAnywhere.exe --version
```

要求：

- 无 Console 窗口
    
- hot path 不启动 PowerShell/cmd
    
- 优先直接调用 Windows API
    
- 已有 symlink 直接复用
    
- 系统修改全部可撤销
    

正式版使用独立 ProgID：

```text
MarkownAnywhere.Markdown
```

不要修改 `Applications\Obsidian.exe`。

---

## 3. Symlink

默认目录：

```text
<Vault>/external-link/
```

命名：

```text
hash(normalized_absolute_path)-filename.md
```

例如：

```text
a81f0921c203-README.md
```

同一源文件重复打开时直接复用。

失效链接定义为：

```text
symlink 存在 + target 不存在
```

默认每 7 天自动清理一次，同时支持插件手动清理。

自动清理不能阻塞文件打开。

---

## 4. 全局配置

配置不能只存在 Vault 内，因为 EXE 需要在 Obsidian 未启动时读取。

位置：

```text
%APPDATA%\MarkownAnywhere\config.json
```

示例：

```json
{
  "version": 1,
  "defaultVault": {
    "name": "Workbench",
    "path": "C:\\Users\\xxx\\Documents\\Obsidian\\workbench"
  },
  "mirrorFolder": "_external-open",
  "cleanupDays": 7
}
```

Plugin 和 EXE 共用此配置。

---

## 5. Obsidian Plugin

插件名称：

```text
Markown Anywhere
```

Plugin ID：

```text
markown-anywhere
```

插件只负责管理，不参与双击 Markdown 的 hot path。

主要功能：

- **Set current vault as default**
    
- Register / Unregister Markdown handler
    
- 设置 symlink 目录
    
- 设置自动清理周期
    
- 查看 Active / Invalid links
    
- 手动清理失效链接
    
- 查看 Companion 状态和版本
    

设置页保持简单：

```text
Markown Anywhere

Default Vault
Current: Workbench
[ Set current vault as default ]

Windows Integration
Status: Registered
[ Register ] [ Unregister ]

External Files
Mirror folder: _external-open
Cleanup interval: 7 days

Symlinks
Active: 18
Invalid: 2
[ Clean invalid links ]

Companion
Status: Connected
Version: 1.0.0
```

Plugin 调用 EXE CLI 获取状态和执行操作，不自行实现 Registry / symlink 逻辑。

---

## 6. 项目结构

采用 GitHub monorepo：

```text
markown-anywhere/
├── companion/
│   ├── cmd/
│   ├── internal/
│   │   ├── config/
│   │   ├── launcher/
│   │   ├── symlink/
│   │   ├── registry/
│   │   └── obsidian/
│   └── go.mod
│
├── plugin/
│   ├── src/
│   ├── manifest.json
│   ├── versions.json
│   └── package.json
│
├── README.md
├── LICENSE
└── .github/workflows/
```

GitHub Repository **只托管源码**，不长期提交编译后的 EXE。

`.gitignore`：

```gitignore
dist/
*.exe
```

---

## 7. GitHub Release 与 Obsidian 插件市场

每个正式版本使用统一版本号，例如：

```text
1.0.0
```

GitHub Release 放编译产物：

```text
Release 1.0.0

├── main.js
├── manifest.json
├── styles.css                  # 可选
└── MarkownAnywhere-Windows-x64.exe
```

职责划分：

```text
GitHub Repository
→ 源码、README、LICENSE、manifest 等

GitHub Release
→ 编译后的 Plugin + Companion EXE

Obsidian Community Plugins
→ 插件目录入口，实际插件文件从 GitHub Release 获取
```

Obsidian 标准安装的是：

```text
main.js
manifest.json
styles.css
```

`MarkownAnywhere.exe` 是额外 Companion，不属于 Obsidian 标准插件文件。

插件可以检测 Companion 是否存在并提供安装说明，但：

- 不自动下载安装 EXE
    
- 不自动更新 EXE
    
- 不自动抢占 `.md` 默认应用
    

系统级操作必须由用户主动触发。

---

## 8. Community Plugin 发布要求

从开发初期就按市场要求设计：

- `isDesktopOnly: true`
    
- README 明确说明：
    
    - 会访问 Vault 外配置
        
    - 会创建 symlink
        
    - Companion 可修改 Windows 文件关联
        
- 默认 local-only
    
- 不上传 Markdown 内容
    
- 不做 telemetry
    
- 不删除用户原始文件
    
- cleanup 只能删除自身创建的 symlink
    
- Companion 缺失时给出明确提示和安装入口
    

---

## 9. v1.0 范围

实现：

- 任意 `.md / .markdown` 双击打开
    
- 默认工作 Vault
    
- Vault 内直接打开
    
- Vault 外 symlink 创建/复用
    
- 中文/Unicode 路径
    
- 同名文件处理
    
- 自动/手动清理失效链接
    
- Windows Register / Unregister
    
- Plugin 设置页面
    
- Set current vault as default
    
- Companion 状态
    
- Symlink 状态管理
    
- GitHub Actions 自动构建与 Release
    

明确不实现：

- 最近文件
    
- 文件历史
    
- 多 Vault 自动路由
    
- 路径规则
    
- 右键高级菜单
    
- 文件副本/同步
    
- 云服务
    
- 后台常驻服务
    
- HTTP/Socket 服务
    
- Telemetry
    

---

## 10. 开发顺序

```text
Phase 1
Go Core
→ config / path / symlink / Obsidian URI

Phase 2
Windows Integration
→ ProgID / register / unregister

Phase 3
Management CLI
→ status / links / cleanup / JSON

Phase 4
Obsidian Plugin
→ Default Vault / Register / Cleanup / Status

Phase 5
Release
→ GitHub Actions / GitHub Release / BRAT Beta

Phase 6
Obsidian Community Plugins submission
```

最终产品定位保持单一：

> **Markown Anywhere：让任意位置的 Markdown 文件都可以直接使用 Obsidian 打开。**