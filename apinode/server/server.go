package server

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	. "com.mgface.disobj/apinode/api"
	. "com.mgface.disobj/apinode/handler"
	. "com.mgface.disobj/common"
)

func StartServer(na, mna, podNamespace string) {
	log.Debug("启动API节点...")
	log.Debug(fmt.Sprintf("节点地址:%s", na))
	log.Debug(fmt.Sprintf("元数据服务节点地址:%s", mna))

	// 创建2个启动标志，一个用来启动发送心跳服务，一个用来更新数据节点数据
	startFlag := make(chan bool, 2)
	go RefreshDynamicMetaNode(mna, podNamespace, startFlag)
	go StartApiHeartbeat(na, startFlag)
	go RefreshDNData(startFlag)

	// 监听请求
	http.HandleFunc("/objects/", ApiHandler)
	http.HandleFunc("/locate/", LocateHandler)

	SupportServeAndGracefulExit(na)
}
