package mgfacemq

import (
	. "com.mgface.disobj/common"
	"com.mgface.disobj/metanode/api"
	"com.mgface.disobj/metanode/cluster"
	"com.mgface.disobj/metanode/mq/mgfacemq/file"
	"com.mgface.disobj/metanode/mq/mgfacemq/memory"
	. "com.mgface.disobj/metanode/mq/mgfacemq/nodeinfo"
	"com.mgface.disobj/metanode/mq/mgfacemq/server"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	//_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"
)

func Startengine(na, ca, gna, ms, pns, svcname string) {

	//启动web端的pprof
	//go http.ListenAndServe(":9909", nil)

	//目录不存在则创建
	if _, err := os.Stat(ms); os.IsNotExist(err) {
		os.MkdirAll(ms, os.ModePerm)
	}

	f, _ := os.OpenFile(fmt.Sprintf("%s%s%s", ms, string(os.PathSeparator), "cpu.prof"),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	pprof.StartCPUProfile(f)

	memoryStore := &memory.MemoryStore{
		StorePath:       ms,
		Mutex:           sync.RWMutex{},
		Datas:           make(map[string]interface{}),
		EnableClean:     true,
		Msgs:            make(chan RecMsg, 1_000),
		FishedSnapshot:  make(chan bool),
		LoadingSnapshot: true,
		BuffMsgs:        make(chan RecMsg, 5_000),
		BuffSemaphore:   make(chan bool),
	}
	serv := &server.Server{
		Store:     memoryStore,
		MutexServ: memoryStore.Mutex,
		Nodeinfo:  &NodeInfo{NodeFlag: "slave", MutexNodeInfo: memoryStore.Mutex},
	}

	//加载内存数据
	file.LoadsnapshotData(memoryStore)

	//启动发送心跳标志
	startflag := make(chan bool)

	//启动集群
	go cluster.StartGossipCluster(na, ca, gna, pns, svcname, serv, startflag)

	//启动心跳服务
	go api.StartMDeartbeat(na, serv, startflag)

	log.Info(fmt.Sprintf("当前PID: 【%d】 ", os.Getpid()), "启动metadataNode...")

	//刷新内存数据和保持内存数据到文件中
	go file.StoresnapshotData(memoryStore)

	//过期内存数据
	go serv.ExpireData(serv.Nodeinfo, 5)

	//启动显示当前内存数据
	go serv.Show()

	rootContext := context.Background()
	//创建一个可以取消的ctx
	ctx, cancelFunc := context.WithCancel(rootContext)
	//启动tcp监听服务
	go serv.Listen(na, ctx)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Debug(<-ch)
	//优雅的停止服务.
	cancelFunc()
	time.Sleep(3 * time.Second)

	pprof.StopCPUProfile()

	log.Info("停止服务.")
}
