# 架构

xpier = 本地 PHP 开发环境管理器(单二进制,~5MB)。核心思路:**无侵入** —— 项目仓库保持干净,一切状态落在 `~/.xpier/`。

## 目录总览

```
internal/
  store/    数据层:类型(Manifest/App/Sites/Site/Lock/AppState...)、JSON/YAML 持久化、
            基础工具(XpierHome、SlugFor、PidAlive、RunOut、PortBusy、SortedKeys、正则...)
  nginx/    配置生成:主配置、站点 server 块、默认 404 站点、ServerPorts/HyperfPort、证书路径、NginxReload
  ca/       本地 CA:生成/信任、*.tld 通配证书、多级域名证书(secure)
  apps/     应用编排:dev.yaml / xpier.yaml apps 的进程管理(up/down/start/restart/log/url/status)
  sites/    站点注册与工具:link/park/unlink/isolate/unisolate/site:php/open/edit/tld/loopback/links/parked
  service/  系统服务:install(sudo)、launchd 守护进程、php-fpm 生命周期、services/service 命令
  share/    隧道:cloudflared 快速隧道管理、状态持久化、shares/share:stop
  xpier/    CLI 分发(run.go)+ 其余命令:init/sync/doctor/status/proxy/db/mail/xdebug/tinker/节点/completion
```

依赖方向(严格单向):

```
store ← nginx ← ca / apps / sites ← service ← share
                      ↑
              xpier(仅分发)
```

`sites → service`(显示 fpm 状态),**禁止** `service → sites`。

## 数据布局(`~/.xpier/`)

| 路径 | 内容 |
|---|---|
| `sites.json` | 站点注册表:TLD、parked 目录、`name -> {path,php,node,driver}` |
| `proxies.json` | 反向代理注册表:`domain -> upstream` |
| `projects/<slug>/` | 每个项目的 `.xpier.yaml`(钉定)+ `.xpier.lock`(sync 解析结果) |
| `nginx/` | nginx.conf、conf.d/*.conf(站点、代理、00-default) |
| `dnsmasq/dnsmasq.conf` | `address=/.test/127.0.0.1` 通配解析 |
| `fpm/` | 每个 PHP 版本的 php-fpm 配置 |
| `certs/` | `wildcard.<tld>.pem` 通配证书、`<domain>.pem` 多级域名证书 |
| `ca/` | 本地 CA(xpier-ca.pem / key) |
| `run/` | unix socket(php-fpm-<ver>.sock) |
| `servers/` | 进程状态 JSON:`fpm-<ver>.json`、`share-<key>.json`、`mailpit.json` |
| `apps/<ns>/` | 应用状态(`<app>.json`)+ 日志(`logs/dev-<app>.log`) |
| `logs/` | nginx / dnsmasq / php-fpm / mailpit / share 日志 |

## 关键流程

### install(一次 sudo)

1. `brew` 安装 nginx + dnsmasq(经真实用户:`sudo -u $SUDO_USER brew`)。
2. 端口自检(80/53 冲突)。
3. 生成 nginx 主配置 + dnsmasq 配置 + 通配证书。
4. 写 `/etc/sudoers.d/xpier`(NOPASSWD 允许 `nginx -s reload/-t`)。
5. 写 `/Library/LaunchDaemons/com.xpier.{nginx,dnsmasq}.plist` 并 bootstrap。
6. 清理旧名守护进程(`com.herdy.*`、`com.pier.*`)。

### link / park → 站点生效

- `link`:注册 `name -> cwd`,生成 nginx server 块(HTTPS 证书、fastcgi→php-fpm socket 或 hyperf proxy_pass),reload。
- `park <dir>`:把目录下的所有子目录同步为站点(`SyncParked`)。
- driver 自动识别:`bin/hyperf.php`→hyperf(proxy)、`public/index.php`→laravel(fastcgi)、`dist/index.html`→spa、否则 static。

### apps 编排(dev.yaml)

```yaml
namespace: devstack          # 进程隔离命名空间
apps:
  php-server:
    dir: /path               # 工作目录
    cmd: php bin/hyperf.php server:watch
    ports: ["9501","9502"]   # 状态/代理用
    php: 8.2                 # 启动前确保版本与扩展
    extensions: [swoole, redis]
  h5:
    dir: /path
    cmd: npm run dev:test
    node: "20"               # 经 nvm 注入 PATH
```

- `up`:逐个启动,端口被占即拒绝;`start <app> [--force]` 单应用;`--force` 会清编译缓存(runtime/container 等,需确认)。
- 进程以独立进程组(`Setpgid`)启动,状态写 `apps/<ns>/<app>.json`;日志 `logs/dev-<app>.log`。
- `url` 按 `app.domain` → `ports[0]` → 状态端口显示。

### share(cloudflared 隧道)

`xpier share <site>`:读取站点,若 `--port N` 则先探测源协议(http/https,兼容 vite basic-ssl),拉起 `cloudflared tunnel --url <target> [--no-tls-verify]`,从日志解析 `*.trycloudflare.com` URL,写 `servers/share-<key>.json`;把隧道 host 追加进站点 server_name 并 reload nginx,使 Host 头能命中。`shares`/`share:stop` 管理。

### secure(CA 与证书)

`sudo xpier secure` → 生成/信任本地 CA(`security add-trusted-cert`),把自签 `*.test` 重签为 CA 签发;`sudo xpier secure <domain>` → 为多级域名(如 `img.test28.test`,通配 `*.test` 覆盖不到)单独签发 SAN 证书。

## 测试策略

单元测试覆盖纯逻辑与状态读写(路径、约束解析、配置生成、注册表往返);exec/sudo/进程类路径不进入单测。详见 docs/development.md。
