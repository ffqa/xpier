package xpier

import "errors"

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
	case "sync":
		return cmdSync(args[1:])
	case "doctor":
		return cmdDoctor(args[1:])
	case "status":
		return cmdStatus(args[1:])
	case "up":
		return cmdUp(args[1:])
	case "down":
		return cmdDown(args[1:])
	case "start":
		return cmdAppStart(args[1:])
	case "restart":
		return cmdAppRestart(args[1:])
	case "log":
		return cmdAppLog(args[1:])
	case "logs":
		return cmdAppLogsAll(args[1:])
	case "url":
		return cmdAppURL(args[1:])
	case "install":
		return cmdInstall(args[1:])
	case "refresh":
		return cmdRefresh(args[1:])
	case "link":
		return cmdLink(args[1:])
	case "unlink":
		return cmdUnlink(args[1:])
	case "park":
		return cmdPark(args[1:])
	case "sites":
		return cmdSites(args[1:])
	case "sites:up":
		return cmdSitesUp(args[1:])
	case "sites:down":
		return cmdSitesDown(args[1:])
	case "site:php":
		return cmdSitePHP(args[1:])
	case "isolate":
		return cmdIsolate(args[1:])
	case "unisolate":
		return cmdUnisolate(args[1:])
	case "isolated":
		return cmdIsolated(args[1:])
	case "php":
		return cmdSitePHPProxy(args[1:])
	case "composer":
		return cmdSiteComposer(args[1:])
	case "debug":
		return cmdSiteDebug(args[1:])
	case "coverage":
		return cmdSiteCoverage(args[1:])
	case "open":
		return cmdOpen(args[1:])
	case "edit":
		return cmdEdit(args[1:])
	case "site-information":
		return cmdSiteInformation(args[1:])
	case "tld":
		return cmdTLD(args[1:])
	case "loopback":
		return cmdLoopback(args[1:])
	case "links":
		return cmdLinks(args[1:])
	case "parked":
		return cmdParked(args[1:])
	case "secure":
		return cmdSecure(args[1:])
	case "secured":
		return cmdSecured(args[1:])
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
		return cmdShare(args[1:])
	case "shares":
		return cmdShares(args[1:])
	case "share:stop":
		return cmdShareStop(args[1:])
	case "mail:up":
		return cmdMailUp(args[1:])
	case "mail:down":
		return cmdMailDown(args[1:])
	case "mail":
		return cmdMail(args[1:])
	case "xdebug":
		return cmdXdebug(args[1:])
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
		return cmdServices(args[1:])
	case "services:stop":
		return cmdServicesStop(args[1:])
	case "services:start":
		return cmdServicesStart(args[1:])
	case "service":
		return cmdService(args[1:])
	case "ini":
		return cmdIni(args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return ErrUsage
	}
}
