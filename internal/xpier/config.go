package xpier

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"xpier/internal/store"
)

// manifestTemplate is the fully-annotated xpier.yaml skeleton: every
// supported field is documented, unused ones stay commented out.
const manifestTemplate = `# xpier 项目钉定配置(可选:没有它 xpier 也能自动检测)
# 所有字段都可选,不需要的直接删掉或保持注释。字段说明见 docs/architecture.md。

runtime: %s                  # 运行时:fpm | hyperf | swoole | frankenphp
# php: "8.4"                # 钉定 PHP 版本(如 "8.2" / "8.4")

# extensions:               # 需要的 PHP 扩展(键=扩展名, 值=版本约束)
#   swoole: "^6.0"          # 约束:"*" 任意 / "^6.0" / ">=6.0"
#   redis: "*"
#   xdebug: "*"
#   # 装法:xpier sync --apply 或 xpier ext:install swoole --php 8.4

# services:                 # 依赖的系统服务(sync --apply 会 brew 安装)
#   - mysql
#   - redis

# apps:                     # 多应用编排(项目内也可用独立 dev.yaml,见 xpier app:init)
#   web:
#     dir: .
#     cmd: php artisan serve
#     port: "8000"
#     domain: web.test      # 可选:生成 nginx 反代
#     node: "20"            # 可选:经 nvm 固定 Node
#     extensions: [swoole]  # 可选:启动前检查扩展
`

// buildManifestContent renders the template with the pinned values uncommented.
func buildManifestContent(php, runtime string) string {
	content := fmt.Sprintf(manifestTemplate, runtime)
	if php != "" {
		content = strings.Replace(content, `# php: "8.4"                # 钉定 PHP 版本(如 "8.2" / "8.4")`,
			fmt.Sprintf(`php: "%s"                # 钉定 PHP 版本`, php), 1)
	}
	return content
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	php := fs.String("php", "", "pin PHP version, e.g. 8.2")
	runtime := fs.String("runtime", "", "runtime (fpm | hyperf | swoole | frankenphp)")
	local := fs.Bool("local", false, "write xpier.yaml into the current directory instead of ~/.xpier (commit it to git if you want it versioned)")
	force := fs.Bool("force", false, "overwrite an existing manifest with the template")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 && fs.Arg(0) == "." {
		*local = true
	}
	if *runtime != "" {
		switch *runtime {
		case "fpm", "hyperf", "swoole", "frankenphp":
		default:
			return fmt.Errorf("invalid runtime %q (fpm | hyperf | swoole | frankenphp)", *runtime)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var manifestPath string
	if *local {
		manifestPath = filepath.Join(cwd, store.ManifestName)
	} else {
		manifestPath, _ = store.ProjectPaths(cwd)
		if err := store.EnsureProjectDir(cwd); err != nil {
			return err
		}
	}
	if _, err := os.Stat(manifestPath); err == nil && !*force {
		return fmt.Errorf("%s already exists (use --force to overwrite with the template)", manifestPath)
	}
	rt := "fpm"
	if *runtime != "" {
		rt = *runtime
	}
	if err := os.WriteFile(manifestPath, []byte(buildManifestContent(*php, rt)), 0o644); err != nil {
		return err
	}
	fmt.Printf("created %s (fully annotated template; unused fields stay commented)\n", manifestPath)
	fmt.Println("manifest is optional: xpier auto-detects runtime and PHP; the manifest only pins them.")
	return nil
}

// cmdInitFresh discards the project's pins and recreates a default manifest
// (Herd's init:fresh).
func cmdInitFresh(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath, lockPath := store.ResolvePaths(cwd)
	os.Remove(manifestPath)
	os.Remove(lockPath)
	if err := store.EnsureProjectDir(cwd); err != nil {
		return err
	}
	if err := os.WriteFile(manifestPath, []byte(buildManifestContent("", "fpm")), 0o644); err != nil {
		return err
	}
	fmt.Printf("recreated %s (annotated defaults; run `xpier init --php 8.2` to re-pin)\n", manifestPath)
	return nil
}
