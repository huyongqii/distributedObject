package command

import (
	. "com.mgface.disobj/common/k8s"
	"com.mgface.disobj/datanode/api"
	"com.mgface.disobj/datanode/server"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/timest/env"
	"github.com/urfave/cli"
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
		Name:     "sdp", //store data path
		Usage:    "数据存储路径(store data path)",
		Required: false,
	},
	//todo 配置文件要单独处理
	cli.StringFlag{
		Name:     "file",
		Usage:    "配置文件",
		Required: false,
	},
}

var StartDNCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"st"},
	Usage: `示例指令:
		startdatanode start -na 127.0.0.1:5000 -mna 127.0.0.1:3001 -sdp C:\\objects
		`,
	Flags: startflag,
	Action: func(ctx *cli.Context) error {
		var nodeaddr, mna, storedatapath, podnamespace string
		//-------解析k8s env
		cfg := new(EnvConfig)
		env.IgnorePrefix()
		err := env.Fill(cfg)
		if err != nil {
			log.Info("未读取到env环境.")
		} else {
			nodeaddr, mna, storedatapath, podnamespace = readEnv(cfg)
		}

		//假如pns为空，说明不是k8s环境
		if podnamespace == "" {
			nodeaddr, mna, storedatapath = readCli(ctx)
		}

		//初始化数据
		api.Initval(storedatapath, nodeaddr)
		//启动服务
		server.StartServer(nodeaddr, mna, podnamespace)
		return nil
	},
}

//读取命令行传入的参数(非k8s使用)
func readCli(ctx *cli.Context) (na, mna, sdp string) {
	//获得节点的地址
	na = ctx.String("na")
	log.Info("读取命令行参数[na]为:", na)

	//获得元数据服务节点地址
	mna = ctx.String("mna")
	log.Info("读取命令行参数[ca]为:", mna)

	//获得节点的数据存储地址
	sdp = ctx.String("sdp")
	log.Info("读取命令行参数[sdp]为:", sdp)
	return
}

func readEnv(cfg *EnvConfig) (na, mna, sdp, pns string) {

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

	//获取数据存储路径
	sdp = cfg.Sdp
	log.Info("读取env[sdp]数据:", sdp)

	//获得节点的命名空检
	pns = cfg.Pns
	log.Info("读取env[pns]数据:", pns)

	return
}
