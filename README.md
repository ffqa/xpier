# xpier

xpier 是 macOS 上的本地 PHP 开发环境管理器(Laravel Herd / Valet 类工具),同时支持多项目应用编排(dev.yaml / xpier.yaml `apps:`)。所有状态都存放在 `~/.xpier/`,**绝不写入你的项目代码**。

## 功能一览(对齐 Herd)

| 领域 | 命令 | 说明 |
|---|---|---|
| 多版本 PHP | `xpier php:use` / `php:list` / `php:install` / `php:update` / `site:up` | 全局默认切换、列表/安装/升级、自动拉起对应 php-fpm |
| nginx 站点 | `xpier site:link`, `site:park`, `site:unlink`, `site:list` | `name.test` 域名,dnsmasq 通配 DNS,自动生成 server 块 |
| 常驻运行时编排 | `xpier app:up / app:down / app:start / app:restart / app:log / app:logs / app:url` | 从 `dev.yaml`/`.xpier.yaml` 的 `apps:` 启动多应用(Hyperf watch、vite 等),端口/进程/日志全托管 |
| 站点工具 | `site:open / site:edit / site:paths / site:which / site:which-php / site:info / site:tld / site:loopback` | 与 Herd 一致的日常命令 |
| PHP 隔离 | `xpier site:isolate / site:unisolate`, `site:php`, `php:exec/composer/debug/coverage` | 按站点固定 PHP 版本,命令代理到站点 PHP |
| 证书 | `sudo xpier secure [domain] / secured` | 本地 CA,签发 `*.test` 通配证书与多级域名证书 |
| 反向代理 | `xpier proxy <domain> <host> / proxies / unproxy` | 代理到任意本地服务(meilisearch、docker 等) |
| 数据库 | `db:install/start/stop`, `svc:available/versions/create`, `db:create`, `db` | MySQL/MariaDB/Redis/Postgres + 内置 Adminer(空密码补丁、自动探测、随机域名) |
| 隧道 | `xpier share [site\|--port N]`, `share:list`, `share:stop` | cloudflared 快速隧道,自动探测 HTTP/HTTPS 源 |
| 邮件 | `xpier mail:up / mail:down / mail` | Mailpit(SMTP 1025, UI 8025) |
| 调试 | `xdebug [status\|on\|off]`, `debug:start / debug:stop`, `php:tinker` | xdebug 开关(+fpm 立即重启)、tinker 自动识别 |
| Node 隔离 | `xpier node:isolate <ver> / node:exec` | 按站点固定 Node 版本(经 nvm) |
| 环境 | `env:init / env:init:fresh / env:sync / doctor / status / svc:status / svc:exec / php:ini / laravel:update` | 版本钉定、依赖解析、体检、服务控制、Laravel 升级 |

## 安装

前置依赖(通过 Homebrew):

```bash
brew install nginx dnsmasq
brew install shivammathur/php/php@8.2   # 按需安装多个版本
```

一次性 root 安装(nginx + dnsmasq launchd 守护进程、证书、sudoers):

```bash
sudo xpier install
sudo xpier secure   # 信任本地 CA
```

## 快速开始

```bash
cd ~/code/my-laravel-app
xpier link                      # 注册为 my-laravel-app.test
open http://my-laravel-app.test

cd ~/code/hyperf-service
xpier isolate 8.2               # 固定 PHP 8.2

# 多应用编排(项目根放 dev.yaml,见 docs/architecture.md)
xpier up
xpier status
```

## 数据与无侵入

- 全部状态: `~/.xpier/`(sites.json、proxies.json、nginx 配置、证书、日志、应用进程状态)
- 项目内只读:链接、park 目录、apps 配置(`dev.yaml` 可选;`xpier init .` 生成隐藏的 `.xpier.yaml` + `.xpier.lock` 钉定/锁定文件)
- 迁移过旧目录名(`~/.herdy`/`~/.pier`)? 执行 `xpier refresh` 重生成 nginx 配置

## 文档

- [docs/architecture.md](docs/architecture.md) — 架构与数据流
- [docs/commands.md](docs/commands.md) — 完整命令参考
- [docs/development.md](docs/development.md) — 构建、测试、发布

## 开发

```bash
make version    # 版本号,由 git commit 数推导(0.0.00,三个段位都逢 99 进位,每次 commit +1)
make build      # 构建 ./xpier(自动注入版本)
make install    # 构建并安装到 /usr/local/bin/xpier
go test ./...   # 全量测试
go vet ./...
```

AGENTS.md 里有给 AI 协作者(以及人)的开发约定,改代码前先读。
