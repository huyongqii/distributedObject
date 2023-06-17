package cluster

import "com.mgface.disobj/metanode/mq/mgfacemq/server"

//启动gossip cluster集群
func StartGossipCluster(nodeAddr, cluster, gossipAddr, podNamespace, serviceName string, serv *server.Server, startflag chan bool) {
	//1.加入集群
	nodename, broadcasts, list := joinGossipCluster(nodeAddr, cluster, gossipAddr, podNamespace, serviceName, serv)

	//2.显示集群状态
	showMemberist(list)

	//3.master发送集群消息
	go sendMsg2Cluster(nodename, serv, broadcasts, list, startflag)
}
