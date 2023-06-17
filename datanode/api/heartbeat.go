package api

import (
	. "com.mgface.disobj/common"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

//启动元数据服务注册

//心跳统一3秒发送一次
func StartDNHeartbeat(nodeAddr string, startflag chan bool) {
	log.Info("获得启动标识:", <-startflag)
	cmd := &Cmd{Name: "set", Key: "dataNodes", Value: nodeAddr}

restart:
	for {
		client := NewReCallFuncTCPClient(GetDNDynamicMetanodeAddr, 3)
		if client == nil {
			log.Warn("datanode心跳包服务连接元数据节点失败，等待重连......")
			goto restart
		}

		//发送心跳包操作
		log.Debug("当前执行的master节点为:", GetDNDynamicMetanodeAddr())
		cmd.Run(client)
		if cmd.Error != nil {
			log.Warn(fmt.Sprintf("%s,datanode心跳包服务发送心跳失败.", time.Now().Format("2006-01-02 15:04:05")))
		}
		time.Sleep(3 * time.Second)
	}
}
