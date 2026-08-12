# 开发指南

## 环境

- Go 1.26+(模块 `xpier`,唯一第三方依赖 `gopkg.in/yaml.v3`)
- macOS;运行时依赖(由 `sudo xpier install` 或用户安装):nginx、dnsmasq、php@X(建议 shivammathur/php)、cloudflared(share)、mailpit(mail)、mysql/postgres(db)

## 构建 / 测试

```bash
go build .             # 产出 ./xpier(约 5MB)
go vet ./...
go test ./...          # 全量单元测试
go test -cover ./...   # 覆盖率
go test ./internal/store/ -run TestManifest -v   # 单测
```

测试原则(也是 AGENTS.md 铁律):

- 每个用例用 `t.Setenv("HOME", t.TempDir())` 隔离 `~/.xpier`,必要时 `os.Chdir` 临时目录并 defer 还原。
- 只测**无副作用**路径:路径/版本/约束/配置生成/注册表往返/状态机。
- **禁止**在单测里执行:`sudo`、启动真实进程(FpmUp/apps up)、打开浏览器、访问外网、`cmdDB`(会下载 Adminer 并 open)。
- macOS `/var` 是 `/private/var` 的符号链接:断言 cwd 时用 `os.Getwd()` 的解析结果。

当前覆盖率:纯逻辑包 57%–86%,总体约 48%。未覆盖部分为 exec/sudo/进程类路径,需要注入命令执行器或集成测试才能打满。

## 包与依赖

```
store ← nginx ← ca / apps / sites ← service ← share
                      ↑
              xpier(仅分发)
```

- 新增领域逻辑:优先放进对应包;`sites → service` 允许,反向禁止(会成环)。
- 命令函数 `func CmdXxx(args []string) error` 在 `internal/xpier/run.go` 的 switch 注册,并同步更新 `usage()` 与 `docs/commands.md`。

## 添加一个命令(示例)

1. 在合适包写 `func CmdExample(args []string) error`(flag 解析、错误带可操作提示)。
2. `run.go` 加 case;`dev_utils.go` 的 completion 列表补上。
3. 写测试(临时 HOME 隔离),`go test ./...`。
4. 更新 `docs/commands.md`。

## 版本规则

版本号 `X.Y.ZZ`(如 `0.0.09`)由 **git commit 数**推导,每次 commit 自动 +1:

- patch = 提交数 % 100,零填充两位
- 三个段位都是 00-99,逢 99 逐级进位:`0.0.99 → 0.1.00 → ... → 0.99.99 → 1.0.00 → ... → 99.99.99`(99 万次提交后回绕到 0.0.00,纯理论)
- `./scripts/version.sh` 打印当前版本;`make version` 同
- 构建/安装统一走 `make build` / `make install`(自动带 `-ldflags "-X xpier/internal/xpier.Version=<版本>"`)
- 手动构建不带 ldflags 时 `xpier --version` 显示 `dev`

## 发布

```bash
make install              # 以当前 commit 数构建并安装到 /usr/local/bin/xpier
VERSION=$(make version)   # 例如 0.0.10

# 通用二进制(Intel + Apple Silicon)
GOOS=darwin GOARCH=amd64 go build -ldflags "-X xpier/internal/xpier.Version=$VERSION" -o dist/xpier-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags "-X xpier/internal/xpier.Version=$VERSION" -o dist/xpier-darwin-arm64 .
lipo -create -output dist/xpier dist/xpier-darwin-amd64 dist/xpier-darwin-arm64
```

发布包包含:README.md、LICENSE、docs/。安装脚本建议:

```bash
sudo install -m 0755 xpier /usr/local/bin/xpier
sudo xpier install
```

## 常见维护点

- 目录改名后(`~/.herdy` → `~/.xpier`):`xpier refresh` 重生成站点配置;install 里也会自动清理旧守护进程。
- nginx reload 走 `/etc/sudoers.d/xpier` 的 NOPASSWD,`NginxReload` 必须带 `-c` 指定 pid 文件。
- Hyperf 编译容器:普通 reload(SIGUSR1)不生效,`restart <app> --force` 才会清 `runtime/container`。
- macOS kqueue 不产生文件编辑事件:watch 类功能依赖 mtime 轮询。
