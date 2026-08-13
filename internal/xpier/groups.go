package xpier

import (
	"fmt"
	"regexp"
	"strings"

	"xpier/internal/store"
)

var usagePadRe = regexp.MustCompile(` {2,}`)

// paintUsageLine colors an indented usage line: command in green, the
// description after the column padding in gray.
func paintUsageLine(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	if !strings.HasPrefix(trimmed, "xpier") {
		return line
	}
	indent := line[:len(line)-len(trimmed)]
	if loc := usagePadRe.FindStringIndex(trimmed); loc != nil {
		cmd := strings.TrimRight(trimmed[:loc[0]], " ")
		desc := trimmed[loc[1]:]
		return indent + store.Green(cmd) + "   " + store.Gray(desc)
	}
	return indent + store.Green(trimmed)
}

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
		"app:status                  project app stack table",
	},
	"site": {
		"site:link [name] [--php 8.4]  link current directory as a site (name.test)",
		"site:park <dir> [...]         serve subdirectories as sites",
		"site:unlink [name] / site:forget   remove a site / forget current dir",
		"site:paths / site:list / site:links / site:parked / site:isolated",
		"site:open [site] / site:edit / site:info <site>",
		"site:tld [x] / site:loopback [x]",
		"site:isolate <ver> / site:unisolate / site:php / site:which / site:which-php",
		"site:up / site:down          start/stop php-fpm for linked sites",
	},
	"php": {
		"php:use [8.4] / php:list / php:install <ver> / php:update [ver]",
		"php:ext <swoole|xdebug|...> [--php 8.4]",
		"php:exec / php:composer / php:debug / php:coverage / php:tinker / php:ini",
	},
	"node": {
		"node:isolate <ver> / node:unisolate / node:isolated",
		"node:exec [--site x] args    run with the site's Node",
	},
	"tls": {
		"secure [domain]             trust CA + sign certs (sudo)",
		"unsecure <site>             serve a site over http only",
		"secured                     list https sites (http-only sites marked)",
	},
	"svc": {
		"svc:status / svc:start / svc:stop   overview / daemons + site fpm",
		"svc:exec <svc> <act>        per-service control (nginx, dnsmasq, php-fpm)",
		"svc:log <svc> / svc:logs [all]      service logs",
		"svc:available / svc:versions / svc:create <svc>",
		"db:install|start|stop <svc> / db:create <name>",
	},
	"config": {
		"site:tld [x] / site:loopback [x]   get/set TLD / loopback",
		"directory-listing [on|off]  toggle nginx autoindex",
		"php:ini [--php 8.4]         open a PHP version's php.ini",
	},
	"env": {
		"env:init [--php 8.4] [.]      pin versions in ~/.xpier/projects",
		"env:init:fresh                reset project pins",
		"env:sync [--apply]            resolve pins; --apply runs brew + writes .xpier.lock",
		"laravel:update                upgrade laravel/framework via composer",
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
		fmt.Println(store.Bold(store.Cyan("xpier groups:")))
		for _, g := range []string{"app", "site", "php", "node", "tls", "svc", "config", "env"} {
			first := namespaceGroups[g][0]
			fmt.Printf("  %s %s\n", store.Green(g), store.Gray(first))
		}
		fmt.Println("  " + store.Gray("more families (global verb + subcommands):"))
		for _, l := range []string{
			"share  / share:list / share:stop      cloudflared tunnel",
			"db     / db:install|start|stop|create  built-in Adminer + databases",
			"mail   / mail:up / mail:down           Mailpit mail capture",
			"debug:start / debug:stop / xdebug      xdebug toggle (immediate fpm restart)",
		} {
			fmt.Println("  " + paintUsageLine("  xpier "+l))
		}
		fmt.Println("  " + store.Gray("run `xpier <group>` for a group's commands, `xpier help` for the full manual"))
		return nil
	}
	if len(args) > 1 {
		return fmt.Errorf("unknown %s subcommand %q; run `xpier %s` for the group's commands", name, args[1], name)
	}
	lines, ok := namespaceGroups[name]
	if !ok {
		return fmt.Errorf("unknown namespace %q (app|site|php|node|tls|svc|config|env|groups)", name)
	}
	fmt.Printf("%s\n", store.Bold(store.Cyan("xpier "+name+": commands")))
	for _, l := range lines {
		fmt.Println(paintUsageLine("  xpier " + l))
	}

	return nil
}
