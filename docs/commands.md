# 命令参考

> `xpier help` 输出同样内容。分组与 run.go 分发保持一致。

## 应用编排(dev.yaml / xpier.yaml apps)

| 命令 | 说明 |
|---|---|
| `xpier app:init [dir] [--force]` | 生成带注释的 dev.yaml 模板 + 操作指南;已有文件需 `--force` 覆盖 |
| `xpier up` / `app:up` | 启动全部 apps(任一已在运行则拒绝) |
| `xpier down` / `app:down` | 停止全部 apps 并清理 nginx 映射 |
| `xpier start <app>` / `app:start` [--force] | 启动单个 app;`--force` = 重启并清编译缓存 |
| `xpier restart <app> [--force]` | kill + 重启单个 app |
| `xpier log <app>` / `app:log` [-f] | 查看单个 app 日志(可 -f 跟随) |
| `xpier logs` | 所有 app 日志一起看 |
| `xpier url [app]` | 显示 app URL |
| `xpier status` | 项目 pins、lock、app 栈表格 |

## 站点(nginx + .test)

| 命令 | 说明 |
|---|---|
| `xpier link [name] [--php 8.2]` | 把当前目录注册为站点(`name.test`);显式名可带 `.def.test` |
| `xpier unlink [name]` | 移除站点 |
| `xpier park <dir> [...]` | 注册目录,其下每个子目录自动成为站点 |
| `xpier sites` | 站点总览(nginx/dnsmasq 状态 + 各站点 fpm 状态) |
| `xpier sites:up / sites:down` | 启动/停止所有站点所需 php-fpm |
| `xpier links` | 列出已链接站点 |
| `xpier parked` | 列出 park 目录 |
| `xpier site:php <site> [version]` | 查看/设置站点 PHP 版本 |
| `xpier isolate <ver> [--site x]` / `unisolate` | 固定/解除当前站点 PHP(与 `site:php` 同效,`isolate` 侧重"固定版本到站点",`site:php` 侧重"查看/设置") |
| `xpier isolated` | 列出已固定 PHP 的站点 |
| `xpier open [site]` / `edit [site]` | 浏览器打开 / IDE 编辑站点 |
| `xpier site-information <site>` | 站点详情(domain/path/driver/php/root) |
| `xpier tld [x]` | 获取/设置 TLD(默认 test) |
| `xpier loopback [x]` | 固定 127.0.0.1 的兼容命令 |

## 证书

| 命令 | 说明 |
|---|---|
| `sudo xpier secure` | 生成并信任本地 CA,签发 `*.test` 通配证书 |
| `sudo xpier secure <domain>` | 为多级域名签发证书(如 `img.test28`) |
| `xpier secured` | 列出全部 https 站点 |

## 反向代理

| 命令 | 说明 |
|---|---|
| `xpier proxy <domain> <host[:port]|http(s)://host:port>` | 代理到任意本地服务(meilisearch、docker...) |
| `xpier proxies` | 列出代理 |
| `xpier unproxy <domain>` | 移除代理 |

## 系统服务

| 命令 | 说明 |
|---|---|
| `sudo xpier install` | 一次性安装:nginx+dnsmasq 守护进程、证书、sudoers、DNS |
| `xpier services` | nginx / dnsmasq / php-fpm / share 总览 |
| `sudo xpier services:start / services:stop` | 启动/停止守护进程(+ 站点 fpm) |
| `xpier service <svc> <act>` | 单服务控制:`nginx|dnsmasq|php-fpm|php-fpm-8.2` × `status|config|configtest|reload|start|stop|restart` |
| `xpier refresh` | 按当前路径重生成全部站点配置并 reload(目录改名/迁移后使用) |

## 环境

| 命令 | 说明 |
|---|---|
| `xpier init [--php 8.2] [--runtime fpm|hyperf|swoole|frankenphp] [.]` | 生成 `~/.xpier/projects/<slug>/xpier.yaml`;`.` 写入仓库内 |
| `xpier sync [--apply]` | 解析钉定(php/扩展/服务);`--apply` 实际执行 brew 并写 `xpier.lock` |
| `xpier doctor` | 体检:PHP 版本、扩展、composer check-platform-reqs |
| `xpier ini [--php 8.2]` | 打开指定版本的 php.ini |
| `xpier completion [bash|zsh]` | 输出 shell 补全 |

## 数据库

| 命令 | 说明 |
|---|---|
| `xpier db:install <svc>` | brew 安装 mysql/mariadb/redis/postgres 并启动 |
| `xpier db:start <svc>` / `db:stop <svc>` | brew services 启停 |
| `xpier db:create <name> [--db mysql]` | 创建数据库 |
| `xpier db [site]` | 打开 Adminer(`database.<tld>`,默认 `database.test`);带 site 时预填该数据库名 |

## 隧道 / 邮件 / 调试

| 命令 | 说明 |
|---|---|
| `xpier share [site\|--port N] [--https]` | cloudflared 隧道(后台托管);`--port` 分享任意本地端口 |
| `xpier shares` | 列出隧道(URL/pid/状态) |
| `xpier share:stop [site]` | 停止隧道 |
| `xpier mail:up / mail:down / mail` | Mailpit:SMTP 127.0.0.1:1025,UI http://127.0.0.1:8025 |
| `xpier xdebug [status\|on\|off] [--php 8.2]` | 切换 xdebug 模式 |

## PHP / Node 工具

| 命令 | 说明 |
|---|---|
| `xpier php [--site x] args...` | 用站点 PHP 执行(透传参数) |
| `xpier composer [--site x] args...` | 站点 composer |
| `xpier debug [--site x]` / `coverage [--site x]` | 带 xdebug/coverage 执行 |
| `xpier tinker [--site x]` | Laravel artisan tinker / Hyperf `bin/hyperf.php tinker` 自动识别 |
| `xpier isolate-node <ver> [--site x]` / `unisolate-node` | 按站点固定 Node(经 nvm) |
| `xpier isolated-node` | 列出固定 Node 的站点 |
| `xpier node [--site x] args...` | 用站点 Node 执行 |

## 其他

| 命令 | 说明 |
|---|---|
| `xpier directory-listing [on\|off]` | 切换 nginx autoindex |
| `xpier forget` | 把当前目录从 parked/站点移除 |
| `xpier fetch-share-url` | (占位)隧道 URL 由 share 进程输出 |
| `xpier help` | 帮助 |
| `xpier --version` / `-v` | 显示版本号(构建时经 `-ldflags "-X xpier/internal/xpier.Version=vX.Y.Z"` 注入) |
