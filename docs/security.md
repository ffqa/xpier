# 安全模型

xpier 是**单机开发工具**,以下设计取舍与风险说明,供使用与审计参考。

## sudo 信任模型

`sudo xpier install` 会写 `/etc/sudoers.d/xpier`,内容为当前用户 `NOPASSWD` 执行**两条精确命令**(不是通配):

- `nginx -s reload -c <~/.xpier/nginx/nginx.conf>`
- `nginx -t -c <~/.xpier/nginx/nginx.conf>`

风险与缓解:

- nginx master 以 root 运行,且 `nginx.conf` 会 `include conf.d/*.conf`(用户可写)。理论上有 conf.d 写权限的用户可以构造配置让 root nginx reload 执行任意指令。
- 这是本地开发工具的必然取舍(用户本就要改站点配置);`~/.xpier` 是当前用户所有。
- 缓解:不要在共享/多用户机器上运行 `sudo xpier install`;配置全部由 xpier 生成,避免手写后不 review。

## 进程安全(防误杀)

- 所有 kill 都先做 **marker 校验**(`store.ProcAlive`):进程存活且 cmdline 含工具写入的标记(`-y <fpm.conf>`、`--url <target>`、`mailpit`、app 的 `Cmd`)才认为是自家进程。
- 重启后 PID 被系统复用不会误杀:状态文件里的旧 PID 无法通过 marker 校验,只会被当作 stale 状态清理。
- 端口级清理只作用于**自家进程组**(pgid),`xpier down` 的孤儿清理只处理 cmdline 匹配 app 命令的进程,其它占用者一律跳过并警告。

## 命令注入防护

- `sync` 的 php 版本(`SafePhpRe`)、扩展名(`safeExtRe`)、服务名(`safeSvcRe`)都校验后才拼进 `brew ...` 命令。
- `proxy` 的 upstream 拒绝 `; \n # { }`,避免注入其它 nginx 指令。
- `node` 透传参数经 `strconv.Quote` 引号化后才进 shell,空格/引号不会被二次解析。

## 数据库凭据

- `db:create` 默认按 brew MySQL/MariaDB 的 socket 认证(空 root 密码)执行;密码受保护的环境用 `--user root --password ...` 显式传入。
- 不要在脚本/CI 里硬编码密码;工具不会持久化数据库密码。

## 证书与密钥

- 本地 CA 私钥与站点私钥在 `~/.xpier/ca/` 与 `~/.xpier/certs/`,权限 0644/目录 0755。
- CA 信任需要 `sudo xpier secure`(写入系统钥匙串)。
- 不要把 `~/.xpier` 分享出去;任何人拿到 CA 私钥都能签发受信任的本地证书。

## 状态文件

- `sites.json` / `proxies.json` / app 状态 / manifest / lock 全部**原子写**(临时文件 + rename),崩溃不会留下截断的注册表。
- 无文件锁:并发运行两个 xpier 命令可能互相覆盖(CLI 场景,风险低;如需要可加锁文件)。

## 非交互行为

- `ConfirmYesNo` 在非 TTY(管道/CI)下自动回复 No:依赖安装类操作会被**静默跳过**,此时 `xpier up` 等会因缺少依赖失败——这是刻意行为,避免 CI 里意外 `brew install`。
