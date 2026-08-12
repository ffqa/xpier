package xpier

import (
	"fmt"
)

func usage() {
	fmt.Println(`xpier - local PHP dev server manager (project files stay in ~/.xpier, never in your repo)

Stack (multi-project orchestration, from dev.yaml / xpier.yaml apps):
  xpier up                              start ALL apps (refuses if any already running)
  xpier start <app> [--force]           start one app (--force = restart with cache clear)
  xpier down                            stop all apps and clean nginx mappings
  xpier restart <app> [--force]         kill + start one app (--force clears compiled caches)
  xpier status                          pins, lock, and app stack table
  xpier log <app> [-f] / xpier logs     one app log / all apps together
  xpier url [app]                       show app URLs

store.Sites:
  xpier link [name] [--php 8.2]         link current directory as a site (name.test)
  xpier park <dir> [...]                serve every subdirectory of a directory as a site
  xpier unlink [name]                   remove a site
  xpier sites / links / parked / isolated   list sites, links, parked dirs, isolated sites
  xpier site:php <site> [version]       show or set a site's PHP version
  xpier isolate <ver> / unisolate       pin/unpin current site's PHP version
  xpier open [site] / edit [site]       open site in browser / IDE
  xpier site-information <site>         show site details
  xpier secure [domain] / secured       trust xpier CA + sign certs / list https sites
  xpier proxy <domain> <host> / proxies / unproxy   reverse proxy to any local service (meilisearch, docker, ...)

Runtime services:
  xpier install                         one-time sudo setup: nginx + dnsmasq launchd daemons, certs, sudoers
  xpier services / service <svc> <act>  service overview / per-service control (nginx, dnsmasq, php-fpm)
  xpier sites:up / sites:down           start/stop php-fpm for linked sites

Environment:
  xpier init [--php 8.2] [--runtime hyperf] [.]  pin versions in ~/.xpier/projects (or '.' in repo)
  xpier sync [--apply]                  resolve pins; --apply runs brew and writes xpier.lock
  xpier doctor                          check environment + composer check-platform-reqs

Extras:
  xpier php/composer/debug/coverage [--site x] args   run with the site's PHP
  xpier db:install|start|stop <svc>     manage MySQL/MariaDB/Redis/Postgres (brew services)
  xpier db:create <name> [--db mysql]   create a database
  xpier db [site]                       open Adminer (database.test)
  xpier share [site|--port N]           tunnel share via cloudflared (managed background)
  xpier shares / share:stop [site]      list tunnels / stop a tunnel
  xpier mail:up / mail:down / mail      Mailpit mail capture (SMTP 1025, UI 8025)
  xpier xdebug [status|on|off] [--php 8.2]  toggle xdebug per PHP version
  xpier tinker / directory-listing / forget   dev utilities
  xpier isolate-node <ver> / node       per-app node version via nvm
  xpier tld [x] / loopback [x]          get/set TLD / loopback
  xpier completion [bash|zsh]           shell completion`)
}
