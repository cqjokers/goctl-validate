package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"goctl-validate/generator"

	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

const Version = "v2.0.0"

var (
	version    = flag.Bool("version", false, "show version and exit")
	help       = flag.Bool("help", false, "show help and exit")
	translator = flag.Bool("translator", false, "generate translator for validation messages")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("goctl-validate %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return
	}

	if *help {
		showHelp()
		return
	}

	// 获取插件上下文
	p, err := plugin.NewPlugin()
	if err != nil {
		fmt.Printf("goctl-validate: %s\n", err)
		os.Exit(1)
	}

	// 检查环境变量
	enableTranslator := *translator
	if os.Getenv("GOCTL_VALIDATE_TRANSLATOR") == "true" {
		enableTranslator = true
	}

	// 使用简化的生成器
	gen := generator.NewSimpleValidateGenerator(p, &generator.Options{
		EnableTranslator: enableTranslator,
	})

	if err := gen.Generate(); err != nil {
		fmt.Printf("goctl-validate: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("goctl-validate: validation code generated successfully")
}

func showHelp() {
	fmt.Println("goctl-validate - A go-zero plugin to generate validation methods")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  goctl api plugin -plugin goctl-validate -api example.api -dir .")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version      show version and exit")
	fmt.Println("  -help         show help and exit")
	fmt.Println("  -translator   generate translator for validation messages (default: false)")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Generates Validate() methods for request structures")
	fmt.Println("  - Uses shared validator instance for better performance")
	fmt.Println("  - Follows go-zero conventions: func (r *Req) Validate() error")
	fmt.Println("  - No modification of existing files")
	fmt.Println("  - Optional Chinese translation support")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Parses API file for structures with validate tags")
	fmt.Println("  2. Generates validate.go in internal/types directory")
	fmt.Println("  3. Creates Validate() method for each structure")
	fmt.Println("  4. Uses shared 'var validate = validator.New()' instance")
}
