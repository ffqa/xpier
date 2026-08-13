package xpier

import (
	"errors"
	"fmt"
	"os"

	"xpier/internal/apps"
	"xpier/internal/ca"
	"xpier/internal/service"
	"xpier/internal/share"
	"xpier/internal/sites"
	"xpier/internal/store"
)

// ErrUsage signals that usage was printed; the caller should exit non-zero
// without printing an extra error line.
var ErrUsage = errors.New("usage")

// legacyCommands maps pre-namespace command names to their canonical form.
// Bare words are GLOBAL commands; project operations carry a namespace
// (site:/app:/php:/node:/env:) and machine services use svc:.
var legacyCommands = map[string]string{
	"up": "app:up", "down": "app:down", "start": "app:start", "restart": "app:restart",
	"url": "app:url", "logs": "svc:logs",
	"link": "site:link", "unlink": "site:unlink", "park": "site:park",
	"forget": "site:forget", "paths": "site:paths", "sites": "site:list",
	"links": "site:links", "parked": "site:parked", "open": "site:open",
	"edit": "site:edit", "site-information": "site:info", "tld": "site:tld",
	"loopback": "site:loopback", "isolate": "site:isolate", "unisolate": "site:unisolate",
	"isolated": "site:isolated", "which": "site:which", "which-php": "site:which-php",
	"sites:up": "site:up", "sites:down": "site:down",
	"use": "php:use", "php": "php:exec", "composer": "php:composer",
	"debug": "php:debug", "coverage": "php:coverage", "tinker": "php:tinker",
	"ini": "php:ini", "ext:install": "php:ext",
	"init": "env:init", "init:fresh": "env:init:fresh", "sync": "env:sync",
	"services": "svc:status", "services:start": "svc:start", "services:stop": "svc:stop",
	"service": "svc:exec", "services:available": "svc:available",
	"services:versions": "svc:versions", "services:create": "svc:create",
	"isolate-node": "node:isolate", "unisolate-node": "node:unisolate",
	"isolated-node": "node:isolated", "node": "node:exec",
	"shares": "share:list",
}

func Run(args []string) error {
	if len(args) < 1 {
		usage()
		return ErrUsage
	}
	if newName, ok := legacyCommands[args[0]]; ok {
		fmt.Fprintf(os.Stderr, "%s: 命令已按命名空间规范迁移,请使用 `xpier %s`(不带前缀的是全局指令;项目指令:site:/app:/php:/node:/env:,本机服务:svc:)\n", store.Red("migrated"), newName)
		return ErrUsage
	}
	switch args[0] {
	// --- 全局指令(不带前缀) ---
	case "status":
		return cmdStatus(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "refresh":
		return cmdRefresh(args[1:])
	case "completion":
		return cmdCompletion(args[1:])
	case "install":
		return service.CmdInstall(args[1:])
	case "xdebug":
		return cmdXdebug(args[1:])
	case "debug:start":
		return cmdDebugStart(args[1:])
	case "debug:stop":
		return cmdDebugStop(args[1:])
	case "db":
		return cmdDB(args[1:])
	case "share":
		return share.CmdShare(args[1:])
	case "mail":
		return cmdMail(args[1:])
	case "fetch-share-url":
		return cmdFetchShareURL(args[1:])
	case "groups", "app", "site", "tls", "ssl", "svc", "config", "env", "php", "node":
		return cmdNamespace(args)

	// --- env: 项目环境钉定 ---
	case "env:init":
		return cmdInit(args[1:])
	case "env:init:fresh":
		return cmdInitFresh(args[1:])
	case "env:sync":
		return cmdSync(args[1:])
	case "laravel:update":
		return cmdLaravelUpdate(args[1:])

	// --- app: 项目应用编排 ---
	case "app:init":
		return apps.CmdInit(args[1:])
	case "app:up":
		return apps.CmdUp(args[1:])
	case "app:down":
		return apps.CmdDown(args[1:])
	case "app:start":
		return apps.CmdStart(args[1:])
	case "app:restart":
		return apps.CmdRestart(args[1:])
	case "app:log":
		return apps.CmdAppLog(args[1:])
	case "app:logs":
		return apps.CmdAppLogsAll(args[1:])
	case "app:url":
		return apps.CmdURL(args[1:])
	case "app:status":
		return apps.CmdStatus(args[1:])

	// --- site: 站点 ---
	case "site:link":
		return sites.CmdLink(args[1:])
	case "site:unlink":
		return sites.CmdUnlink(args[1:])
	case "site:park":
		return sites.CmdPark(args[1:])
	case "site:forget":
		return cmdForget(args[1:])
	case "site:paths":
		return sites.CmdPaths(args[1:])
	case "site:list":
		return sites.CmdSites(args[1:])
	case "site:links":
		return sites.CmdLinks(args[1:])
	case "site:parked":
		return sites.CmdParked(args[1:])
	case "site:open":
		return sites.CmdOpen(args[1:])
	case "site:edit":
		return sites.CmdEdit(args[1:])
	case "site:info":
		return sites.CmdSiteInformation(args[1:])
	case "site:tld":
		return sites.CmdTLD(args[1:])
	case "site:loopback":
		return sites.CmdLoopback(args[1:])
	case "site:isolate":
		return sites.CmdIsolate(args[1:])
	case "site:unisolate":
		return sites.CmdUnisolate(args[1:])
	case "site:isolated":
		return sites.CmdIsolated(args[1:])
	case "site:php":
		return sites.CmdSitePHP(args[1:])
	case "site:which":
		return sites.CmdWhich(args[1:])
	case "site:which-php":
		return sites.CmdWhichPHP(args[1:])
	case "site:up":
		return sites.CmdSitesUp(args[1:])
	case "site:down":
		return sites.CmdSitesDown(args[1:])

	// --- php: PHP 运行时 ---
	case "php:use":
		return sites.CmdUse(args[1:])
	case "php:list":
		return service.CmdPhpList(args[1:])
	case "php:install":
		return service.CmdPhpInstall(args[1:])
	case "php:update":
		return service.CmdPhpUpdate(args[1:])
	case "php:ext":
		return service.CmdExtInstall(args[1:])
	case "php:exec":
		return sites.CmdSitePHPProxy(args[1:])
	case "php:composer":
		return sites.CmdSiteComposer(args[1:])
	case "php:debug":
		return sites.CmdSiteDebug(args[1:])
	case "php:coverage":
		return sites.CmdSiteCoverage(args[1:])
	case "php:tinker":
		return cmdTinker(args[1:])
	case "php:ini":
		return cmdIni(args[1:])

	// --- svc: 本机服务 ---
	case "svc:status":
		return service.CmdServices(args[1:])
	case "svc:start":
		return service.CmdServicesStart(args[1:])
	case "svc:stop":
		return service.CmdServicesStop(args[1:])
	case "svc:available":
		return service.CmdServicesAvailable(args[1:])
	case "svc:versions":
		return service.CmdServicesVersions(args[1:])
	case "svc:create":
		return service.CmdServicesCreate(args[1:])
	case "svc:log":
		return apps.CmdLog(args[1:])
	case "svc:logs":
		return apps.CmdLogsAll(args[1:])
	case "svc:exec":
		return service.CmdService(args[1:])

	// --- node: Node 隔离与执行 ---
	case "node:isolate":
		return cmdIsolateNode(args[1:])
	case "node:unisolate":
		return cmdUnisolateNode(args[1:])
	case "node:isolated":
		return cmdIsolatedNode(args[1:])
	case "node:exec":
		return cmdNode(args[1:])

	// --- db: / mail: / share: ---
	case "db:install":
		return cmdDBInstall(args[1:])
	case "db:start":
		return cmdDBStart(args[1:])
	case "db:stop":
		return cmdDBStop(args[1:])
	case "db:create":
		return cmdDBCreate(args[1:])
	case "share:list":
		return share.CmdShares(args[1:])
	case "share:stop":
		return share.CmdShareStop(args[1:])
	case "mail:up":
		return cmdMailUp(args[1:])
	case "mail:down":
		return cmdMailDown(args[1:])

	// --- 证书与代理 ---
	case "secure":
		return ca.CmdSecure(args[1:])
	case "secured":
		return ca.CmdSecured(args[1:])
	case "unsecure":
		return sites.CmdUnsecure(args[1:])
	case "proxy":
		return cmdProxy(args[1:])
	case "proxies":
		return cmdProxies(args[1:])
	case "unproxy":
		return cmdUnproxy(args[1:])
	case "directory-listing":
		return cmdDirectoryListing(args[1:])

	case "help", "-h", "--help":
		usage()
		return nil
	case "--version", "-v":
		fmt.Printf("xpier %s\n", Version)
		return nil
	default:
		usage()
		return ErrUsage
	}
}
