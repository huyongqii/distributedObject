package cluster

import (
	. "com.mgface.disobj/common"
	"com.mgface.disobj/metanode/mq/mgfacemq/server"
	"encoding/base64"
	"encoding/json"
	"github.com/hashicorp/memberlist"
	"sync"
)

type delegate struct {
	mtx        sync.RWMutex
	items      map[string]interface{}
	broadcasts *memberlist.TransmitLimitedQueue
	serv       *server.Server
}

type gossipInfo struct {
	Action string // add, del
	Data   interface{}
}

func buildGossipInfo(Action string, Data interface{}) gossipInfo {
	return gossipInfo{
		Action: Action,
		Data:   Data,
	}
}

func (proxy *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

//获取gossip服务传递过来的数据
func (proxy *delegate) NotifyMsg(data []byte) {
	dst := make([]byte, len(data))
	copy(dst, data)
	if len(dst) == 0 {
		return
	}

	var ginfo gossipInfo
	json.Unmarshal(dst, &ginfo)

	decoded, _ := base64.StdEncoding.DecodeString(ginfo.Data.(string))
	dx, _ := GzipDecode(decoded)
	if ginfo.Action == "heartbeat" { //心跳数据
		proxy.serv.AssignThis(dx)
	}
}

func (proxy *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return proxy.broadcasts.GetBroadcasts(overhead, limit)
}

func (proxy *delegate) LocalState(join bool) []byte {
	//proxy.mtx.RLock()
	//m := proxy.items
	//proxy.mtx.RUnlock()
	//data, _ := json.Marshal(m)
	//return data
	return nil
}

func (proxy *delegate) MergeRemoteState(buf []byte, join bool) {
	//if len(buf) == 0 {
	//	return
	//}
	//if !join {
	//	return
	//}
	//var m map[string]string
	//if err := json.Unmarshal(buf, &m); err != nil {
	//	return
	//}
	//proxy.mtx.Lock()
	//for k, v := range m {
	//	proxy.items[k] = v
	//}
	//proxy.mtx.Unlock()
}
