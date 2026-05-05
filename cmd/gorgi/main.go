// Command gorgi 是项目可执行入口，仅做参数解析与错误打印，不承担任何业务逻辑。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/holosola/gorgi/internal/app"
)

func main() {
	cfgPath := flag.String("c", "", "配置文件路径，留空则按 GORGI_CONFIG -> ./configs/config.yaml 顺序查找")
	flag.Parse()

	if err := app.Run(*cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "gorgi 启动失败: %v\n", err)
		os.Exit(1)
	}
}
