package cluster

import (
	"com.mgface.disobj/common"
	"com.mgface.disobj/common/k8s"
	. "com.mgface.disobj/metanode/api"
	"com.mgface.disobj/metanode/mq/mgfacemq/server"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/memberlist"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//加入gossip集群
func joinGossipCluster(nodeAddr, cluster, gossipAddr, podNamespace, serviceName string, serv *server.Server) (string, *memberlist.TransmitLimitedQueue, *memberlist.Memberlist) {
	//假如长度为1，说明是从k8s yaml env环境过来的值，那么gaddr是一个端口值
	splitInfo := strings.Split(gossipAddr, ":")
	//如果是k8s，gossipAddr只有端口
	gossipPort := gossipAddr
	if len(splitInfo) == 1 {
		//截取nodeAddr的IP地址，不要端口
		gossipAddr = fmt.Sprintf("%s:%s", strings.Split(nodeAddr, ":")[0], gossipAddr)
	}
	conf := memberlist.DefaultLANConfig()
	proxy := &delegate{
		mtx:   sync.RWMutex{},
		items: make(map[string]interface{}),
		serv:  serv,
	}
	conf.Delegate = proxy
	conf.Events = &mgfaceEventDelegate{}
	conf.UDPBufferSize = 50_000 //gossip包传输的最大长度
	//以节点启动的纳秒数作为节点名称
	currentNano := strconv.FormatInt(time.Now().UnixNano(), 10)
	nodename := fmt.Sprintf("%s-%s", currentNano, nodeAddr)
	//节点名称为：节点创建的纳秒数-节点的addr值
	log.Info("当前节点名称:", nodename)
	conf.Name = nodename
	splitInfo = strings.Split(gossipAddr, ":")
	if len(splitInfo) != 2 {
		log.Fatal("元数据IP地址格式为IP:Port")
		return "", nil, nil
	}
	//截取IP地址
	conf.BindAddr = splitInfo[0]
	//截取PORT端口
	conf.BindPort, _ = strconv.Atoi(splitInfo[1])
	conf.LogOutput = ioutil.Discard
	list, err := memberlist.Create(conf)
	if err != nil {
		panic("错误创建集群信息: " + err.Error())
	}
	broadcasts := &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return list.NumMembers()
		},
		RetransmitMult: 1, //最大传输次数
	}
	proxy.broadcasts = broadcasts
	//加入存在的集群节点，最少需要指定一个已知的节点信息
	allNodes := strings.Split(cluster, ",")

	//假如长度为1，说明是从k8s yaml env环境过来的值
	if len(allNodes) == 1 {
		allNodes = k8s.DetectMetaService(gossipPort, podNamespace, serviceName)
	}

	list.Join(allNodes)
	if err != nil {
		panic("错误加入集群信息: " + err.Error())
	}
	return nodename, broadcasts, list
}

//显示当前集群信息
func showMemberist(list *memberlist.Memberlist) {
	go func() {
		for {
			log.Debug("@@@@@@@@@当前集群@@@@@@@@@")
			for i, v := range list.Members() {
				log.Debug(fmt.Sprintf("节点(%d)-%s", i, v.Name))
			}
			log.Debug("#######10S~15S刷新#######")
			time.Sleep(time.Duration(10+rand.Intn(5)) * time.Second)
		}
	}()
}

//用来启动心跳服务
var startHBRun sync.Once

//给gossip集群发送消息
func sendMsg2Cluster(nodename string, serv *server.Server, broadcasts *memberlist.TransmitLimitedQueue,
	list *memberlist.Memberlist, startflag chan bool) {
	//todo 1.如果存在网络分区，会存在多个master,这个需要重新调整代码,要考虑合并多个master数据存储元数据（心跳数据不需要）
	//todo 2.如果重新连接上之后发现自己状态从master变成slave之后，那么需要把数据同步给当前的master

	//定义打印计数器
	printCount := 0
	for {
		wnode := WrapMemberlistNodes(list.Members())
		sort.Stable(wnode)
		if serv.Nodeinfo.DecideMaster() {
			goto breakMaster
		}

		if wnode[0].Name == nodename {
			serv.Nodeinfo.SetMaster()
			log.Info(fmt.Sprintf("把当前节点[%s]设置为master.", nodename))
		}
	breakMaster:
		//当前节点是master，那么只有它才能写数据给其他节点
		if serv.Nodeinfo.DecideMaster() {
			//只写心跳数据给其他metanode节点
			strbytes := []byte(serv.ThisToJson())
			//log.Debug("压缩前长度:", len(strbytes))
			zipbyte, _ := common.GzipEncode(strbytes)
			//log.Debug("压缩后长度:", len(zipbyte))
			encoded := base64.StdEncoding.EncodeToString(zipbyte)

			gossipInfo := buildGossipInfo("heartbeat", encoded)

			data, _ := json.Marshal(gossipInfo)

			//广播数据
			broadcasts.QueueBroadcast(&broadcast{
				msg:    data,
				notify: nil,
			})
		}

		//取排序之后第一个节点为master
		//wnode[0].Name数据格式为:节点创建的纳秒数-节点的addr值
		masterMetadata := strings.Split(wnode[0].Name, "-")[1]
		if printCount > 10 {
			log.Info("当前节点状态:", serv.Nodeinfo.GetNodeInfo())
			log.Debug("masterMetadata::", masterMetadata)
			printCount = 0
		}

		//假如master有变化，那么直接重新同步一次
		if !serv.Nodeinfo.DecideMaster() && masterMetadata != GetDynamicMNAddr() {
			log.Info(fmt.Sprintf("当前master:%s,上一个master:%s", masterMetadata, GetDynamicMNAddr()))
			//发送sync同步请求master的快照文件，告知接收快照的服务端口是哪一个
			info := make(map[string]string)
			info["length"] = "all"
			info["nodeinfo"] = nodename
			dx, _ := json.Marshal(info)
			//log.Debug("压缩前长度:", len(strbytes))
			zipbyte, _ := common.GzipEncode(dx)
			//log.Debug("压缩后长度:", len(zipbyte))
			data := base64.StdEncoding.EncodeToString(zipbyte)
			//添加客户端服务IP到缓冲发送里面
			client := common.NewReconTCPClient(masterMetadata, 3)
			datax := []byte(fmt.Sprintf("X%d %s", len(data), data))
			client.Conn.Write(datax)

			recode := make([]byte, 4096)
			n, _ := client.Conn.Read(recode)
			replyMsg := string(recode[:n])
			log.Debug("replyMsg::", replyMsg)
		}

		SetDynamicMNAddr(masterMetadata)
		//启动心跳服务
		startHBRun.Do(func() {
			startflag <- true
			close(startflag)
		})
		time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)
		printCount++
	}
}
