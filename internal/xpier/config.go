package xpier

import (
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

runtime: %s                  # 项目类型/运行时(决定 xpier up 怎么处理):
#  ┌ fpm        标准 PHP / Laravel —— 无需 up:cd 项目 && xpier link,nginx+php-fpm 直接服务,访问 <目录名>.test
#  ├ static     静态站点            —— 同 fpm,link 后 nginx 直接出文件
#  ├ hyperf     Hyperf 常驻服务     —— 在下方 apps: 写启动命令,xpier up 拉起
#  ├ swoole     Swoole 常驻服务     —— 同 hyperf
#  └ frankenphp FrankenPHP          —— 同 hyperf
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

// parseInitArgs parses init flags position-independently (Go's flag package
// stops at the first positional arg, so `xpier init . --force` would silently
// drop --force).
func parseInitArgs(args []string) (php, runtime string, local, force bool, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			fmt.Println("usage: xpier init [--php 8.2] [--runtime fpm|hyperf|swoole|frankenphp] [--local] [--force] [.]")
			return "", "", false, false, nil
		case a == "--php" && i+1 < len(args):
			php = args[i+1]
			i++
		case strings.HasPrefix(a, "--php="):
			php = strings.TrimPrefix(a, "--php=")
		case a == "--runtime" && i+1 < len(args):
			runtime = args[i+1]
			i++
		case strings.HasPrefix(a, "--runtime="):
			runtime = strings.TrimPrefix(a, "--runtime=")
		case a == "--local":
			local = true
		case a == "--force":
			force = true
		case a == ".":
			local = true
		default:
			return "", "", false, false, fmt.Errorf("unexpected argument %q", a)
		}
	}
	return php, runtime, local, force, nil
}

func cmdInit(args []string) error {
	php, runtime, local, force, err := parseInitArgs(args)
	if err != nil {
		return err
	}
	if runtime != "" {
		switch runtime {
		case "fpm", "static", "hyperf", "swoole", "frankenphp":
		default:
			return fmt.Errorf("invalid runtime %q (fpm | static | hyperf | swoole | frankenphp)", runtime)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var manifestPath string
	if local {
		manifestPath = filepath.Join(cwd, store.ManifestName)
	} else {
		manifestPath, _ = store.ProjectPaths(cwd)
		if err := store.EnsureProjectDir(cwd); err != nil {
			return err
		}
	}
	if _, err := os.Stat(manifestPath); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite with the template)", manifestPath)
	}
	rt := runtime
	if rt == "" {
		rt = "fpm"
	}
	if err := os.WriteFile(manifestPath, []byte(buildManifestContent(php, rt)), 0o644); err != nil {
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
