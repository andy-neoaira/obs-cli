package main

// main.go 是 obs-cli 的入口文件。
// Go 程序的执行总是从 main 包中的 main() 函数开始。
// 这里我们将所有逻辑委托给 cmd 包，保持入口极简。

import (
	"os"

	"github.com/andy-neoaira/obs-cli/cmd"
)

var exitProcess = os.Exit

func main() {
	// Execute 负责解析命令行参数并分发到对应的子命令
	exitProcess(cmd.Execute())
}
