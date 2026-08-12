package xpier

import (
	"fmt"
	"strings"
	"xpier/internal/store"
)

func usage() {
	s := fmt.Sprintf(`xpier %s - local PHP dev server manager (project files stay in ~/.xpier, never in your repo)

Apps (multi-project orchestration, from dev.yaml / xpier.yaml apps):
  xpier app / app:init [dir] [--force]   group help / dev.yaml template
  xpier up / app:up                      start ALL apps (refuses if any already running)
  xpier start <app> / app:start [--force]   start one app
  xpier down / app:down                  stop all apps and clean nginx mappings
  xpier restart <app> / app:restart [--force]   kill + start one app
  xpier status                           pins, lock, and app stack table
  xpier app:log <app> [-f] / app:logs   project app log / all project app logs
  xpier log <svc> [-f]                  one service log (nginx, dnsmasq, php-fpm, mailpit)
  xpier logs [all]                      service logs (global; all = + project app logs)
  xpier url [app] / app:url              show app URLs

Sites & Projects:
  xpier link [name] [--php 8.2]          link current directory as a site (name.test)
  xpier park <dir> [...]             serve subdirectories as sites
  xpier unlink [name] / forget           remove a site / forget the current directory
  xpier paths / sites / links / parked   list project paths / sites / links / parked dirs
  xpier open [site] / edit [site]        open site in browser / IDE
  xpier site-information <site>          show site details
  xpier db [db] [--site name]            open built-in Adminer (auto-detects MySQL)

PHP Version Management:
  xpier use [8.3]                        show / set the global default PHP version
  xpier php:list / php:install <ver>     list installed PHP / install one via brew
  xpier isolate <ver> / unisolate        pin/unpin current site's PHP version
  xpier isolated                         list PHP-isolated sites
  xpier which / which-php                current site's PHP version / binary path
  xpier site:php <site> [version]        show or set a site's PHP version
  xpier php/composer [--site x] args     run with the site's PHP
  xpier ini [--php 8.2]                  open a PHP version's php.ini

Node.js Version Management:
  xpier isolate-node <ver> / unisolate-node / isolated-node   per-site Node via nvm
  xpier node [--site x] args             run with the site's Node

SSL/TLS:
  xpier secure [domain] / unsecure <site> / secured   sign certs / http-only / list

Proxies:
  xpier proxy <domain> <host> / proxies / unproxy   reverse proxy to any local service

Services:
  xpier install                  one-time sudo setup (daemons, certs, sudoers)
  xpier services / services:start|stop   service overview / start-stop daemons + site fpm
  xpier service <svc> <act>              per-service control (nginx, dnsmasq, php-fpm)
  xpier sites:up / sites:down            start/stop php-fpm for linked sites
  xpier db:install|start|stop <svc>   manage MySQL/MariaDB/Redis/Postgres
  xpier db:create <name> [--db mysql]    create a database
  xpier mail:up / mail:down / mail       Mailpit mail capture (SMTP 1025, UI 8025)

Debugging:
  xpier debug / coverage [--site x]      run with xdebug / coverage enabled
  xpier xdebug [status|on|off] [--php 8.2]   toggle xdebug per PHP version
  xpier tinker [--site x]                Laravel / Hyperf tinker (auto-detected)

Sharing:
  xpier share [site|--port N]        cloudflared tunnel (managed)
  xpier shares / share:stop [site]       list tunnels / stop a tunnel
  xpier fetch-share-url                  print the tunnel URL of a running share

Configuration:
  xpier tld [x] / loopback [x]           get/set TLD / loopback
  xpier directory-listing [on|off]       toggle nginx autoindex

Environment & Tooling:
  xpier init [--php 8.2] [.]      pin versions in ~/.xpier/projects
  xpier sync [--apply]                   resolve pins; --apply runs brew and writes xpier.lock
  xpier doctor                           check environment + composer check-platform-reqs
  xpier refresh                          regenerate all configs after a home move
  xpier completion [bash|zsh]            shell completion
  xpier --version / -v                   show version`, Version)
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
