package main

import (
	"com.mgface.disobj/common"
	. "com.mgface.disobj/metanode/command"
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
 *   元数据服务主要提供存储元信息
 * 
 *
`

func main() {
	//执行的终端命令
	cmds := []cli.Command{
		StartMNCommand,
	}
	common.RunFn(usage, cmds)
}
