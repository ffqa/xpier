package xpier

import "fmt"

// namespaceGroups maps a bare namespace word to its subcommand usage lines.
// The Run dispatcher prints the group when the bare word is typed and no
// concrete command matches (e.g. `xpier app` lists all app:* commands).
// Words that are already real verbs (php, node, db, share, mail, proxy,
// debug, log) stay verbs; their group is visible via `xpier groups` / help.
var namespaceGroups = map[string][]string{
	"app": {
		"app:init [dir] [--force]    generate a commented dev.yaml template + guide",
		"app:up                      start ALL apps (refuses if any already running)",
		"app:down                    stop all apps and clean nginx mappings",
		"app:start <app> [--force]   start one app (--force = restart with cache clear)",
		"app:restart <app> [--force] kill + start one app",
		"app:log <app> [-f]          one app log",
		"app:logs                    all app logs together",
		"app:url [app]               show app URLs",
		"status                      pins, lock, and app stack table",
	},
	"site": {
		"link [name] [--php 8.2]     link current directory as a site (name.test)",
		"park <dir> [...]            serve every subdirectory of a directory as a site",
		"unlink [name] / forget      remove a site / forget the current directory",
		"paths / sites / links / parked   list project paths / sites / links / parked dirs",
		"open [site] / edit [site]   open site in browser / IDE",
		"site-information <site>     show site details",
		"db [db] [--site name]       open built-in Adminer (auto-detects MySQL)",
	},
	"tls": {
		"secure [domain]             trust CA + sign certs (sudo)",
		"unsecure <site>             serve a site over http only",
		"secured                     list https sites (http-only sites marked)",
	},
	"svc": {
		"install                     one-time sudo setup: daemons, certs, sudoers",
		"services / services:start|stop   overview / start-stop daemons + site fpm",
		"service <svc> <act>         per-service control (nginx, dnsmasq, php-fpm)",
		"sites:up / sites:down       start/stop php-fpm for linked sites",
		"db:install|start|stop <svc> manage MySQL/MariaDB/Redis/Postgres",
		"db:create <name> [--db mysql]   create a database",
		"services:available / services:versions / services:create <svc>",
		"mail:up / mail:down / mail  Mailpit mail capture",
	},
	"config": {
		"tld [x] / loopback [x]      get/set TLD / loopback",
		"directory-listing [on|off]  toggle nginx autoindex",
		"ini [--php 8.2]             open a PHP version's php.ini",
	},
	"env": {
		"init [--php 8.2] [...]      pin versions in ~/.xpier/projects (or '.' in repo)",
		"init:fresh                  reset project pins",
		"sync [--apply]              resolve pins; --apply runs brew + writes xpier.lock",
		"doctor                      check environment + composer check-platform-reqs",
		"refresh                     regenerate all configs after a home move",
		"laravel:update              upgrade laravel/framework via composer",
		"completion [bash|zsh]       shell completion",
	},
}

// cmdNamespace prints the command group for a bare namespace word. The word
// "groups" prints the full group overview.
func cmdNamespace(args []string) error {
	name := "app"
	if len(args) > 0 {
		name = args[0]
	}
	if name == "groups" {
		fmt.Println("xpier groups:")
		for _, g := range []string{"app", "site", "tls", "svc", "config", "env"} {
			fmt.Printf("  %-8s %s\n", g, namespaceGroups[g][0])
		}
		fmt.Println("  also: php (PHP group), node (Node group), debug/log (Debugging), share (Sharing), proxy (Proxies), db/mail (Services)")
		fmt.Println("run `xpier <group>` for a group's commands, `xpier help` for the full manual")
		return nil
	}
	lines, ok := namespaceGroups[name]
	if !ok {
		return fmt.Errorf("unknown namespace %q (app|site|tls|svc|config|env|groups)", name)
	}
	fmt.Printf("xpier %s: commands\n", name)
	for _, l := range lines {
		fmt.Printf("  xpier %s\n", l)
	}
	if name == "app" {
		fmt.Println("flat aliases: xpier up/start/down/restart/log/logs/url (no prefix)")
	}
	return nil
}
