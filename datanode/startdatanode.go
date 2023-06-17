package main

import (
	"com.mgface.disobj/common"
	. "com.mgface.disobj/datanode/command"
	"github.com/urfave/cli"
)

const usage = `
 *      ┌─┐       ┌─┐
 *   ┌──┘ ┴───────┘ ┴──┐
 *   │       ───       │
 *   │  ─┬┘       └┬─  │
 *   │       ─┴─       │
 *   └───┐         ┌───┘
 *       │         └─────────┐
 *       │                   ├─┐
 *       │                   ┌─┘
 *       └─┐  ┐  ┌──┬──┐  ┌──┘
 *         │ ─┤ ─┤  │ ─┤ ─┤
 *         └──┴──┘  └──┴──┘
 *   神兽保佑
 *   代码无BUG!
 *
 *   数据服务主要提供存储对象数据
 * 
 *
`

func main() {
	//执行的终端命令
	cmds := []cli.Command{
		StartDNCommand,
	}
	common.RunFn(usage, cmds)
}
