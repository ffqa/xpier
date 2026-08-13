# 命令参考

> 命名规范:**不带前缀的是全局指令**(本机级);项目指令一律带命名空间(`app:`/`site:`/`php:`/`node:`/`env:`);本机服务用 `svc:`。旧裸命令运行时会提示迁移后的名字。

## 全局指令(不带前缀)

| 命令 | 说明 |
|---|---|
| `xpier status` | 服务总览(nginx/dnsmasq/php-fpm/share)+ 站点数 |
| `xpier doctor` | 环境体检 + 钉定/lock 摘要 + composer check-platform-reqs |
| `xpier refresh` | 迁移目录后重生成全部配置 |
| `sudo xpier install` | 一次性安装:nginx+dnsmasq 守护进程、证书、sudoers |
| `xpier xdebug [status\|on\|off] [--php 8.4]` | 切换 xdebug |
| `xpier debug:start` / `debug:stop` | 开关 xdebug 并立即重启 php-fpm |
| `xpier db [db] [--site name]` | 打开内置 Adminer(自动探测 MySQL,默认名被占用时随机域名) |
| `xpier share [site\|--port N]` | cloudflared 隧道(后台托管) |
| `xpier share:list` / `share:stop [site]` | 列出 / 停止隧道 |
| `xpier mail` / `mail:up` / `mail:down` | Mailpit(SMTP 1025,UI 8025) |
| `xpier secure [domain]` / `unsecure <site>` / `secured` | 证书签发 / http-only / 列表 |
| `xpier proxy <domain> <host>` / `proxies` / `unproxy` | 反向代理 |
| `xpier directory-listing [on\|off]` | 切换 nginx autoindex |
| `xpier groups` / `<app\|site\|php\|node\|tls\|svc\|config\|env>` | 分组总览 / 组内命令 |
| `xpier completion [bash\|zsh]` / `--version` / `-v` | 补全 / 版本 |

## 项目 — env(钉定)

| 命令 | 说明 |
|---|---|
| `xpier env:init [--php 8.4] [--runtime fpm\|static\|hyperf\|swoole\|frankenphp] [--force] [.]` | 生成 `~/.xpier/projects/<slug>/.xpier.yaml`;`.` 写入仓库内隐藏文件 |
| `xpier env:init:fresh` | 丢弃钉定,重建默认 manifest |
| `xpier env:sync [--apply]` | 解析钉定;`--apply` 执行 brew(自动 trust tap)并写 `.xpier.lock` |
| `xpier laravel:update` | 站点内 `composer update laravel/framework --with-all-dependencies` |

## 项目 — app(应用编排,dev.yaml / .xpier.yaml apps)

| 命令 | 说明 |
|---|---|
| `xpier app:init [dir] [--force]` | 生成带注释 dev.yaml 模板 + 操作指南 |
| `xpier app:up` / `app:down` | 启动/停止栈(网站型条目自动 link) |
| `xpier app:start <app>` / `app:restart <app>` [--force] | 单个应用;`--force` 清编译缓存 |
| `xpier app:log <app> [-f]` / `app:logs` | 项目应用日志 |
| `xpier app:url [app]` / `app:status` | 应用 URL / 栈表格(网站型显示 `site`) |

## 项目 — site(站点)

| 命令 | 说明 |
|---|---|
| `xpier site:link [name] [--php 8.4]` | 链接当前目录(`name.test`),自动拉起 php-fpm |
| `xpier site:unlink [name]` / `site:forget` | 移除站点 / 忘记当前目录 |
| `xpier site:park <dir> [...]` | 目录下每个子目录自动成站 |
| `xpier site:list` / `site:links` / `site:parked` / `site:isolated` / `site:paths` | 各种列表 |
| `xpier site:open [site]` / `site:edit` / `site:info <site>` | 浏览器 / IDE / 详情 |
| `xpier site:tld [x]` / `site:loopback [x]` | 获取/设置 TLD / loopback |
| `xpier site:isolate <ver>` / `site:unisolate` / `site:php <site> [ver]` | 站点 PHP 固定 |
| `xpier site:which` / `site:which-php` | 当前站点 PHP 版本 / 二进制路径 |
| `xpier site:up` / `site:down` | 启动/停止所有站点 php-fpm |

## 项目 — php / node

| 命令 | 说明 |
|---|---|
| `xpier php:use [8.4]` | 查看/设置全局默认 PHP |
| `xpier php:list` / `php:install <ver>` / `php:update [ver]` | 列出 / 安装 / 升级 PHP |
| `xpier php:ext <swoole\|xdebug\|...> [--php 8.4]` | 安装扩展(自动 trust tap) |
| `xpier php:exec` / `php:composer` / `php:debug` / `php:coverage` / `php:tinker` / `php:ini` | 用站点 PHP 执行 |
| `xpier node:isolate <ver>` / `node:unisolate` / `node:isolated` | 站点 Node 固定(经 nvm) |
| `xpier node:exec [--site x] args` | 用站点 Node 执行 |

## 本机 — svc(服务)

| 命令 | 说明 |
|---|---|
| `xpier svc:status` | nginx/dnsmasq/php-fpm/share 总览 |
| `sudo xpier svc:start` / `svc:stop` | 守护进程 + 站点 fpm |
| `xpier svc:exec <nginx\|dnsmasq\|php-fpm> <status\|config\|configtest\|reload\|start\|stop\|restart>` | 单服务控制 |
| `xpier svc:log <svc>` / `svc:logs [all]` | 服务日志;`all` 叠加项目 app 日志 |
| `xpier svc:available` / `svc:versions` | 可装服务 / 已装版本 |
| `xpier svc:create <mysql\|mariadb\|redis\|postgres\|mailpit>` | 安装并启动服务 |
| `xpier db:install\|start\|stop <svc>` / `db:create <name> [--db mysql] [--user --password]` | 数据库管理 |

## 旧命令迁移对照(运行旧名会自动提示)

`up→app:up`、`down→app:down`、`start→app:start`、`restart→app:restart`、`log→svc:log`、`logs→svc:logs`、`url→app:url`、`link→site:link`、`unlink→site:unlink`、`park→site:park`、`forget→site:forget`、`paths→site:paths`、`sites→site:list`、`links→site:links`、`parked→site:parked`、`open→site:open`、`edit→site:edit`、`site-information→site:info`、`tld→site:tld`、`loopback→site:loopback`、`isolate→site:isolate`、`unisolate→site:unisolate`、`isolated→site:isolated`、`which→site:which`、`which-php→site:which-php`、`sites:up→site:up`、`sites:down→site:down`、`use→php:use`、`php→php:exec`、`composer→php:composer`、`debug→php:debug`、`coverage→php:coverage`、`tinker→php:tinker`、`ini→php:ini`、`ext:install→php:ext`、`init→env:init`、`init:fresh→env:init:fresh`、`sync→env:sync`、`services→svc:status`、`services:start→svc:start`、`services:stop→svc:stop`、`service→svc:exec`、`services:available→svc:available`、`services:versions→svc:versions`、`services:create→svc:create`、`isolate-node→node:isolate`、`unisolate-node→node:unisolate`、`isolated-node→node:isolated`、`node→node:exec`、`shares→share:list`
