package xpier

import "fmt"

// namespaceGroups maps a bare namespace name to its subcommand usage lines.
// The Run dispatcher prints the group when the bare name is typed and no
// concrete command matches (e.g. `xpier app` lists all app:* commands).
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
	},
}

// cmdNamespace prints the command group for a bare namespace name. It returns
// an error for namespaces that have no registered group.
func cmdNamespace(args []string) error {
	name := "app"
	if len(args) > 0 {
		name = args[0]
	}
	lines, ok := namespaceGroups[name]
	if !ok {
		return fmt.Errorf("unknown namespace %q", name)
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
