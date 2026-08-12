# AGENTS.md — xpier 仓库协作者须知

本文件给 AI 代理与人类协作者:在改动本仓库前,先读这里。

## 铁律(不可违反)

1. **绝不修改用户项目代码。** xpier 是无侵入工具,所有写入都必须在 `~/.xpier/`(或测试的临时 HOME)内。禁止编辑、创建、删除任何 `dev.yaml` 之外的项目文件,禁止改用户配置文件(vite.config.ts、php.ini 等)。
2. **未经明确许可,不得停止/重启用户的进程。** 测试、示例、调试都不得调用 `apps.CmdUp/CmdDown/CmdRestart`、`service.FpmUp/FpmDown`、`share.CmdShare` 等会启动/杀掉真实进程的路径。
3. **不要运行带真实副作用的命令做"验证"。** 例如 `cmdDB` 会下载 Adminer 并 `open` 浏览器;`CmdOpen` 会打开浏览器。测试一律用临时 HOME 隔离,且只测无副作用路径。

## 构建与测试

```bash
go build .            # 产出 ./xpier(约 5MB)
go vet ./...
go test ./...         # 全量单元测试
go test -cover ./...  # 覆盖率
```

- 测试必须可隔离:用 `t.Setenv("HOME", t.TempDir())` 重置数据目录,必要时 `os.Chdir` 到临时目录并 `defer` 还原。
- macOS `/var` → `/private/var` 是符号链接:比较 cwd 路径时用 `os.Getwd()` 的解析结果,不要直接用 `t.TempDir()` 返回值。
- 不要写会执行 `sudo`、启动真实进程、访问网络的测试(除非显式标注集成测试)。

## 包结构(按领域拆分)

```
main.go                    入口,仅调用 xpier.Run
internal/
  store/                   数据层:类型、持久化、基础工具(唯一被所有包依赖)
  nginx/                   nginx 配置生成、端口/证书路径、reload
  ca/                      本地 CA、证书签发、secure/secured
  apps/                    多应用编排:dev.yaml/xpier.yaml apps 的 up/down/start/restart/log/url
  sites/                   站点注册:link/park/unlink/isolate/tld 及站点工具
  service/                 install 安装、launchd 守护进程、php-fpm 管理、services 命令
  share/                   cloudflared 隧道
  xpier/                   CLI 分发(run.go)+ 其余命令(init/sync/doctor/status/proxy/db/extras)
```

依赖方向:**store ← nginx ← ca/apps/sites ← service ← share**;`sites` 与 `service` 之间是 `sites → service`(sites 用 `service.FpmRunning` 显示状态),不允许反向依赖。`xpier` 只做命令分发,不承载领域逻辑。

## 约定

- 命令函数签名统一为 `func CmdXxx(args []string) error`,在 `internal/xpier/run.go` 的 switch 里注册。
- 写文件的命令如果目标目录可能不存在,必须先 `os.MkdirAll`。
- 与执行环境强相关的命令(install 需 root、fpm 需真实 php-fpm)失败时给可操作的错误提示(带 `brew install ...` 命令)。
- 注释只解释 WHY;迁移后不要留下"// store.UDPBusy reports..." 这类挪动残留注释。
- 提交粒度:一个领域一个 commit;commit message 用英文,概括 WHY。

## 数据位置

全部状态在 `~/.xpier/`:`sites.json`、`proxies.json`、`projects/<name>/xpier.yaml+xpier.lock`、`nginx/`、`dnsmasq/`、`fpm/`、`certs/`、`ca/`、`run/`(sock+pid)、`servers/`(fpm/share/mailpit 状态 json)、`apps/<ns>/`(应用状态与日志)、`logs/`。
