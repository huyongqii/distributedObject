package command

import (
	"com.mgface.disobj/apinode/api"
	"com.mgface.disobj/apinode/server"
	. "com.mgface.disobj/common/k8s"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/timest/env"
	"github.com/urfave/cli"
	"strconv"
)

//todo 1.读取配置文件，覆盖默认值

//todo 2.根据配置文件来初始化文件存储还是缓存存储数据

//todo 3.判断如果是启用文件存储，启动的时候需要加载已有的文件数据

var startflag = []cli.Flag{
	cli.StringFlag{
		Name:     "na", //node address
		Usage:    "节点地址(node addres)",
		Required: false,
	},
	cli.StringFlag{
		Name:     "mna", //metanode address
		Usage:    "元数据服务节点地址(metanode address).只要配置一个可以正常连接上的种子节点即可",
		Required: false,
	},
	cli.StringFlag{
		Name:     "ds", //datashards number
		Usage:    "数据分片大小数量(datashards number)",
		Required: false,
	},
	cli.StringFlag{
		Name:     "ps", //parityshards number
		Usage:    "奇偶校验数量(parityshards number)",
		Required: false,
	},

	//todo 配置文件要单独处理
	cli.StringFlag{
		Name:     "file",
		Usage:    "配置文件",
		Required: false,
	},
}

var ApiStartCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"st"},
	Usage: `示例指令:
		startapinode start -na 127.0.0.1:6000 -mna 127.0.0.1:3001 -ds 2 -ps 1
		`,
	Flags: startflag,
	Action: func(ctx *cli.Context) error {
		var nodeaddr, mna, podnamespace string
		var ds, ps int
		//-------解析k8s env
		cfg := new(EnvConfig)
		env.IgnorePrefix()
		err := env.Fill(cfg)
		if err != nil {
			log.Info("未读取到env环境.")
		} else {
			nodeaddr, mna, ds, ps, podnamespace = readEnv(cfg)
		}
		//假如pns为空，说明不是k8s环境
		if podnamespace == "" {
			nodeaddr, mna, ds, ps = readCli(ctx)
		}

		//初始化数据
		api.Initval(ds, ps)
		//启动服务
		server.StartServer(nodeaddr, mna, podnamespace)
		return nil
	},
}

// 读取命令行传入的参数(非k8s使用)
func readCli(ctx *cli.Context) (na, mna string, ds, ps int) {
	//获得节点的地址
	na = ctx.String("na")
	log.Info("读取命令行参数[na]为:", na)

	//获得元数据服务节点地址
	mna = ctx.String("mna")
	log.Info("读取命令行参数[ca]为:", mna)

	//获得数据分片数量
	ds = ctx.Int("ds")
	log.Info("读取命令行参数[ds]为:", ds)

	//获得奇偶校验分片数量
	ps = ctx.Int("ps")
	log.Info("读取命令行参数[ds]为:", ps)
	return
}

func readEnv(cfg *EnvConfig) (na, mna string, ds, ps int, pns string) {

	//从env获取节点的地址
	na = cfg.Na
	log.Info("读取env[na]数据:", na)

	//从env获取到端口
	napt := cfg.Napt
	log.Info("读取env[napt]数据:", napt)
	if napt != "" {
		na = fmt.Sprintf("%s:%s", na, napt)
	}
	//从env获取到集群地址
	ca := cfg.Ca
	log.Info("读取env[ca]数据:", ca)
	//从env获取到集群地址的默认端口
	capt := cfg.Capt
	log.Info("读取env[capt]数据:", capt)

	mna = fmt.Sprintf("%s:%s", ca, capt)
	log.Info("metanode: ", mna)

	//获取数据分片大小数量
	ds, _ = strconv.Atoi(cfg.Ds)
	log.Info("读取env[ds]数据:", ds)

	//获取奇偶校验数量
	ps, _ = strconv.Atoi(cfg.Ps)
	log.Info("读取env[ps]数据:", ps)

	//获得节点的命名空检
	pns = cfg.Pns
	log.Info("读取env[pns]数据:", pns)

	return
}
