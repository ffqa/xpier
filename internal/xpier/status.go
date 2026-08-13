package xpier

import (
	"fmt"

	"xpier/internal/service"
	"xpier/internal/store"
)

// cmdStatus is the GLOBAL overview: services + linked sites. The project
// app stack lives under `xpier app:status`.
func cmdStatus(args []string) error {
	service.XpierServiceStatus()
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	fmt.Printf("sites:   %d linked (`xpier site:list` for details)\n", len(sites.Sites))
	return nil
}
