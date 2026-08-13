package xpier

import (
	"fmt"
	"strings"

	"xpier/internal/store"
)

func usage() {
	s := fmt.Sprintf(`xpier %s - local PHP dev server manager (project files stay in ~/.xpier, never in your repo)

GLOBAL (no namespace = machine-wide):
  xpier status                            services + linked sites overview
  xpier doctor                            environment check + pins + lock
  xpier refresh                           regenerate all configs after a home move
  xpier install                           one-time sudo setup (daemons, certs, sudoers)
  xpier xdebug [status|on|off] [--php 8.4] / debug:start|stop    toggle xdebug
  xpier db [db] [--site name]             open built-in Adminer (auto-detects MySQL)
  xpier share [site|--port N] / share:list / share:stop          cloudflared tunnel
  xpier mail / mail:up / mail:down        Mailpit mail capture (SMTP 1025, UI 8025)
  xpier secure [domain] / unsecure <site> / secured / proxy ...  TLS + reverse proxy
  xpier groups / <app|site|tls|svc|config|env|php|node>         group listings
  xpier completion [bash|zsh] / --version / -v / help

PROJECT - env (pins):
  xpier env:init [--php 8.4] [--runtime fpm|static|hyperf|swoole|frankenphp] [--force] [.]
  xpier env:init:fresh                    reset project pins
  xpier env:sync [--apply]                resolve pins; --apply runs brew + writes .xpier.lock
  xpier laravel:update                    upgrade laravel/framework via composer

PROJECT - app (stack, from dev.yaml / .xpier.yaml apps):
  xpier app:init [dir] [--force]          generate a commented dev.yaml template + guide
  xpier app:up / app:down                 start / stop the stack (web-type entries auto-link)
  xpier app:start <app> / app:restart <app> [--force]   one app
  xpier app:log <app> [-f] / app:logs     project app logs
  xpier app:url [app] / app:status        app URLs / stack table

PROJECT - site:
  xpier site:link [name] [--php 8.4]      link current directory (name.test)
  xpier site:unlink [name] / site:forget  remove a site / forget current dir
  xpier site:park <dir> / site:paths      park dirs / list project paths
  xpier site:list / site:links / site:parked / site:isolated
  xpier site:open [site] / site:edit / site:info <site>
  xpier site:tld [x] / site:loopback [x]  get/set TLD / loopback
  xpier site:isolate <ver> / site:unisolate / site:php <site> [ver]
  xpier site:which / site:which-php       current site's PHP version / binary
  xpier site:up / site:down               start/stop php-fpm for linked sites

PROJECT - php / node:
  xpier php:use [8.4] / php:list / php:install <ver> / php:update [ver]
  xpier php:ext <swoole|xdebug|...> [--php 8.4] / php:ini [--php 8.4]
  xpier php:exec|composer|debug|coverage|tinker [--site x] args
  xpier node:isolate <ver> / node:unisolate / node:isolated / node:exec [--site x] args

MACHINE - svc (services):
  xpier svc:status / svc:start / svc:stop  overview / daemons + site fpm
  xpier svc:exec <nginx|dnsmasq|php-fpm> <status|config|configtest|reload|start|stop|restart>
  xpier svc:log <svc> / svc:logs [all]    service logs (nginx, dnsmasq, php-fpm, mailpit)
  xpier svc:available / svc:versions / svc:create <mysql|mariadb|redis|postgres|mailpit>
  xpier db:install|start|stop <svc> / db:create <name> [--db mysql]

CONFIG:
  xpier directory-listing [on|off]        toggle nginx autoindex`, Version)
	var b strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		t := strings.TrimRight(line, " ")
		switch {
		case i == 0:
			b.WriteString(store.Bold(t))
		case t != "" && !strings.HasPrefix(t, " ") && strings.HasSuffix(t, ":") && !strings.HasPrefix(t, "xpier"):
			b.WriteString(store.Bold(store.Cyan(t)))
		case strings.HasPrefix(line, "  "):
			b.WriteString(paintUsageLine(line))
		default:
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	fmt.Print(b.String())
}
