package api

import (
	. "com.mgface.disobj/common"
	"com.mgface.disobj/metanode/mq/mgfacemq/server"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

//metadata节点也需要注册，供Apinode和datanode使用
//
//nodeAddr 节点当前的值
//
//nodeflag  节点当前是master还是slave
func StartMDeartbeat(nodeAddr string, serv *server.Server, startflag chan bool) {
	log.Info("获得启动标识:", <-startflag)
restart:
	for {
		client := NewReCallFuncTCPClient(GetDynamicMNAddr, 3)
		if client == nil {
			log.Warn("metanode心跳包服务连接master节点失败，等待重连......")
			goto restart
		}
		//发送心跳包操作
		log.Debug("当前执行的master节点为:", GetDynamicMNAddr())
		metanodeInfo := fmt.Sprintf("%s-%s", nodeAddr, serv.Nodeinfo.GetNodeInfo())
		cmd := &Cmd{Name: "set", Key: "metaNodes", Value: metanodeInfo}
		cmd.Run(client)
		if cmd.Error != nil {
			log.Warn(fmt.Sprintf("%s,metanode心跳包服务发送心跳失败.", time.Now().Format("2006-01-02 15:04:05")))
		}
		time.Sleep(3 * time.Second)
	}
}
