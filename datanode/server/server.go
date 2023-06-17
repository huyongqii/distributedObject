package server

import (
	. "com.mgface.disobj/common"
	. "com.mgface.disobj/datanode/api"
	"com.mgface.disobj/datanode/datarepair"
	. "com.mgface.disobj/datanode/hander"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func StartServer(na, mna, podnamespace string) {
	log.Info("启动数据节点...")
	log.Info(fmt.Sprintf("节点地址:%s", na))
	log.Info(fmt.Sprintf("元数据服务节点地址:%s", mna))
	//后台数据修复
	go datarepair.Repair()

	//创建2个启动标志，一个用来启动发送心跳服务，一个用来更新数据节点数据

	startflag := make(chan bool)

	//更新dynamicMetanodeAddr
	go RefreshDNMetanodeAddr(mna, podnamespace, startflag)
	//心跳汇报
	go StartDNHeartbeat(na, startflag)

	http.HandleFunc("/objects/", ApiHandler)

	SupportServeAndGracefulExit(na)
}
