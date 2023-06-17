package main

import (
	. "com.mgface.disobj/apinode/command"
	"com.mgface.disobj/common"
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
 *   API服务主要提供对象数据对外服务
 * 
 *
`

func main() {
	//执行的终端命令
	cmds := []cli.Command{
		ApiStartCommand,
	}
	common.RunFn(usage, cmds)
}
