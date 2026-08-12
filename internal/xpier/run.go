package xpier

import (
	"errors"
	"fmt"

	"xpier/internal/apps"
	"xpier/internal/ca"
	"xpier/internal/service"
	"xpier/internal/share"
	"xpier/internal/sites"
)

// ErrUsage signals that usage was printed; the caller should exit non-zero
// without printing an extra error line.
var ErrUsage = errors.New("usage")

func Run(args []string) error {
	if len(args) < 1 {
		usage()
		return ErrUsage
	}
	switch args[0] {
	case "init":
		return cmdInit(args[1:])
	case "init:fresh":
		return cmdInitFresh(args[1:])
	case "laravel:update":
		return cmdLaravelUpdate(args[1:])
	case "sync":
		return cmdSync(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "status":
		return cmdStatus(args[1:])
	case "app", "site", "tls", "ssl", "svc", "config", "env", "groups":
		return cmdNamespace(args[1:])
	case "app:init":
		return apps.CmdInit(args[1:])
	case "up", "app:up":
		return apps.CmdUp(args[1:])
	case "down", "app:down":
		return apps.CmdDown(args[1:])
	case "start", "app:start":
		return apps.CmdStart(args[1:])
	case "restart", "app:restart":
		return apps.CmdRestart(args[1:])
	case "log":
		return apps.CmdLog(args[1:])
	case "logs":
		return apps.CmdLogsAll(args[1:])
	case "app:log":
		return apps.CmdAppLog(args[1:])
	case "app:logs":
		return apps.CmdAppLogsAll(args[1:])
	case "url", "app:url":
		return apps.CmdURL(args[1:])
	case "install":
		return service.CmdInstall(args[1:])
	case "refresh":
		return cmdRefresh(args[1:])
	case "paths":
		return sites.CmdPaths(args[1:])
	case "which":
		return sites.CmdWhich(args[1:])
	case "which-php":
		return sites.CmdWhichPHP(args[1:])
	case "use":
		return sites.CmdUse(args[1:])
	case "unsecure":
		return sites.CmdUnsecure(args[1:])
	case "php:list":
		return service.CmdPhpList(args[1:])
	case "php:install":
		return service.CmdPhpInstall(args[1:])
	case "php:update":
		return service.CmdPhpUpdate(args[1:])
	case "link":
		return sites.CmdLink(args[1:])
	case "unlink":
		return sites.CmdUnlink(args[1:])
	case "park":
		return sites.CmdPark(args[1:])
	case "sites":
		return sites.CmdSites(args[1:])
	case "sites:up":
		return sites.CmdSitesUp(args[1:])
	case "sites:down":
		return sites.CmdSitesDown(args[1:])
	case "site:php":
		return sites.CmdSitePHP(args[1:])
	case "isolate":
		return sites.CmdIsolate(args[1:])
	case "unisolate":
		return sites.CmdUnisolate(args[1:])
	case "isolated":
		return sites.CmdIsolated(args[1:])
	case "php":
		return sites.CmdSitePHPProxy(args[1:])
	case "composer":
		return sites.CmdSiteComposer(args[1:])
	case "debug":
		return sites.CmdSiteDebug(args[1:])
	case "coverage":
		return sites.CmdSiteCoverage(args[1:])
	case "open":
		return sites.CmdOpen(args[1:])
	case "edit":
		return sites.CmdEdit(args[1:])
	case "site-information":
		return sites.CmdSiteInformation(args[1:])
	case "tld":
		return sites.CmdTLD(args[1:])
	case "loopback":
		return sites.CmdLoopback(args[1:])
	case "links":
		return sites.CmdLinks(args[1:])
	case "parked":
		return sites.CmdParked(args[1:])
	case "secure":
		return ca.CmdSecure(args[1:])
	case "secured":
		return ca.CmdSecured(args[1:])
	case "proxy":
		return cmdProxy(args[1:])
	case "proxies":
		return cmdProxies(args[1:])
	case "unproxy":
		return cmdUnproxy(args[1:])
	case "db:install":
		return cmdDBInstall(args[1:])
	case "db:start":
		return cmdDBStart(args[1:])
	case "db:stop":
		return cmdDBStop(args[1:])
	case "db:create":
		return cmdDBCreate(args[1:])
	case "db":
		return cmdDB(args[1:])
	case "share":
		return share.CmdShare(args[1:])
	case "shares":
		return share.CmdShares(args[1:])
	case "share:stop":
		return share.CmdShareStop(args[1:])
	case "mail:up":
		return cmdMailUp(args[1:])
	case "mail:down":
		return cmdMailDown(args[1:])
	case "mail":
		return cmdMail(args[1:])
	case "xdebug":
		return cmdXdebug(args[1:])
	case "debug:start":
		return cmdDebugStart(args[1:])
	case "debug:stop":
		return cmdDebugStop(args[1:])
	case "tinker":
		return cmdTinker(args[1:])
	case "directory-listing":
		return cmdDirectoryListing(args[1:])
	case "forget":
		return cmdForget(args[1:])
	case "isolate-node":
		return cmdIsolateNode(args[1:])
	case "unisolate-node":
		return cmdUnisolateNode(args[1:])
	case "isolated-node":
		return cmdIsolatedNode(args[1:])
	case "node":
		return cmdNode(args[1:])
	case "completion":
		return cmdCompletion(args[1:])
	case "fetch-share-url":
		return cmdFetchShareURL(args[1:])
	case "services":
		return service.CmdServices(args[1:])
	case "services:available":
		return service.CmdServicesAvailable(args[1:])
	case "services:versions":
		return service.CmdServicesVersions(args[1:])
	case "services:create":
		return service.CmdServicesCreate(args[1:])
	case "services:stop":
		return service.CmdServicesStop(args[1:])
	case "services:start":
		return service.CmdServicesStart(args[1:])
	case "service":
		return service.CmdService(args[1:])
	case "ini":
		return cmdIni(args[1:])
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
