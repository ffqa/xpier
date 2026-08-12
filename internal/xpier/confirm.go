package xpier

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirmYesNo asks the user for confirmation (default no).
func confirmYesNo(prompt string) (bool, error) {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	}
	return false, nil
}

// ensureBrewPackage prompts to install a missing brew package and installs it
// on confirmation. Returns nil when the binary is already present.
func ensureBrewPackage(bin, formula, display string) error {
	if fileExists(bin) {
		return nil
	}
	ok, err := confirmYesNo(fmt.Sprintf("%s 未安装（brew install %s），是否现在安装？", display, formula))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%s not installed; run `brew install %s` and retry", display, formula)
	}
	fmt.Printf("installing %s...\n", formula)
	if out, err := brewAsUser("install", formula); err != nil {
		return fmt.Errorf("brew install %s: %v: %s", formula, err, out)
	}
	return nil
}
