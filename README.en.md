# xpier

[中文](README.md) | English

xpier is a local PHP development environment manager for **macOS** — a CLI tool in the same spirit as Laravel Herd / Valet — with built-in multi-app orchestration (`dev.yaml` / `xpier.yaml` `apps:`). All state lives in `~/.xpier/` and it **never writes into your project code**.

> **Platform note:** xpier currently runs on **macOS only** (both Apple Silicon and Intel).
> There is **no GUI yet** — xpier is a pure command-line tool, and can be used as a drop-in CLI replacement for Laravel Herd's command line.

## Features (aligned with Herd)

| Area | Commands | Description |
|---|---|---|
| Multi-version PHP | `xpier php:use` / `php:list` / `php:install` / `php:update` / `site:up` | switch the global default, list/install/upgrade versions, auto-start the matching php-fpm |
| nginx sites | `xpier site:link`, `site:park`, `site:unlink`, `site:list` | `name.test` domains, wildcard DNS via dnsmasq, server blocks generated for you |
| Runtime orchestration | `xpier app:up / app:down / app:start / app:restart / app:log / app:logs / app:url` | start multiple apps (Hyperf watch, vite, ...) from `dev.yaml`/`.xpier.yaml` `apps:` — ports, processes and logs fully managed |
| Site tools | `site:open / site:edit / site:paths / site:which / site:which-php / site:info / site:tld / site:loopback` | the everyday commands, matching Herd |
| PHP isolation | `xpier site:isolate / site:unisolate`, `site:php`, `php:exec/composer/debug/coverage` | pin a PHP version per site; commands are proxied to the site's PHP |
| Certificates | `sudo xpier secure [domain] / secured` | local CA that issues `*.test` wildcard and multi-level domain certificates |
| Reverse proxy | `xpier proxy <domain> <host> / proxies / unproxy` | proxy any local service (meilisearch, docker, ...) onto a `.test` domain |
| Databases | `db:install/start/stop`, `svc:available/versions/create`, `db:create`, `db` | MySQL/MariaDB/Redis/Postgres + built-in Adminer (empty-password patch, auto-detection, random domain) |
| Tunnels | `xpier share [site\|--port N]`, `share:list`, `share:stop` | quick cloudflared tunnels, auto-detecting HTTP/HTTPS sources |
| Mail | `xpier mail:up / mail:down / mail` | Mailpit (SMTP 1025, UI 8025) |
| Debugging | `xdebug [status\|on\|off]`, `debug:start / debug:stop`, `php:tinker` | xdebug toggles (+ immediate fpm restart), tinker auto-detection |
| Node isolation | `xpier node:isolate <ver> / node:exec` | pin a Node version per site (via nvm) |
| Environment | `env:init / env:init:fresh / env:sync / doctor / status / svc:status / svc:exec / php:ini / laravel:update` | version pinning, dependency resolution, health check, service control, Laravel upgrades |

## Installation

### 1. Prerequisites (via Homebrew)

```bash
brew install nginx dnsmasq
brew install shivammathur/php/php@8.2   # install as many PHP versions as you need
```

### 2. Get the binary

Download the latest release for your Mac from [GitHub Releases](https://github.com/ffqa/xpier/releases):

- `xpier-<version>-darwin-arm64` — Apple Silicon (M1/M2/M3/M4)
- `xpier-<version>-darwin-amd64` — Intel Macs

Then make it executable and put it on your `PATH`:

```bash
chmod +x xpier-*-darwin-*
sudo mv xpier-*-darwin-* /usr/local/bin/xpier
```

Or build from source:

```bash
git clone git@github.com:ffqa/xpier.git
cd xpier
make install        # builds and installs to /usr/local/bin/xpier
```

### 3. One-time root install

Sets up nginx + dnsmasq launchd daemons, certificates and sudoers:

```bash
sudo xpier install
sudo xpier secure   # trust the local CA
```

## Quick start

```bash
cd ~/code/my-laravel-app
xpier link                      # registers as my-laravel-app.test
open http://my-laravel-app.test

cd ~/code/hyperf-service
xpier isolate 8.2               # pin PHP 8.2 for this site

# Multi-app orchestration (put a dev.yaml in your project root, see docs/architecture.md)
xpier up
xpier status
```

## Data & non-invasive design

- All state lives in `~/.xpier/` (sites.json, proxies.json, nginx configs, certificates, logs, app process state)
- Inside your project everything is read-only: links, parked directories, apps config (`dev.yaml` is optional; `xpier init .` generates a hidden `.xpier.yaml` + `.xpier.lock` pinning/lock file)
- Migrated from an old data directory (`~/.herdy`/`~/.pier`)? Run `xpier refresh` to regenerate nginx configs

## Documentation

- [docs/architecture.md](docs/architecture.md) — architecture & data flow
- [docs/commands.md](docs/commands.md) — full command reference
- [docs/development.md](docs/development.md) — build, test, release

## Development

```bash
make version    # version derived from git commit count (0.0.00, each segment carries at 99, +1 per commit)
make build      # builds ./xpier (version injected automatically)
make install    # builds and installs to /usr/local/bin/xpier
go test ./...   # full test suite
go vet ./...
```

See AGENTS.md for contributor conventions (for both AI agents and humans) — read it before changing code.
