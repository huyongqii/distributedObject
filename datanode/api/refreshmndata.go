package api

import (
	. "com.mgface.disobj/common"
	"com.mgface.disobj/common/k8s"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
	"strings"
	"time"
)

// 更新dynamicMetanodeAddr值，每500毫秒更新一次
func RefreshDNMetanodeAddr(metanodeAddr, podnamespace string, startflag chan bool) {
	if podnamespace != "" {
		val := strings.Split(metanodeAddr, ":")
		metanodes := k8s.DetectMetaService(val[1], podnamespace, val[0])
		if len(metanodes) > 0 {
			//随机选一个metanode
			metanodeAddr = metanodes[0]
		}
		log.Debug("随机选取的metanodeAddr:", metanodeAddr)
	}
	//该metanodeAddr作为种子节点取元数据
	client := NewReconTCPClient(metanodeAddr, 3)
	if client == nil {
		log.Fatal(fmt.Sprintf("元数据服务连接失败,请提供一个正确的元数据服务IP地址"))
		os.Exit(-1)
	}

	//第一次查询提供元数据服务的信息
	cmd := &Cmd{Name: "get", Key: "metaNodes", Value: ""}

	//为了得到master节点信息
	cmd.Run(client)
	var results []MetaValue
	json.Unmarshal([]byte(cmd.Value), &results)
	for _, metavale := range results {
		//当key为metaNodes，RealNodeValue值为:"nodeaddr-(master|slave)"
		data := strings.Split(metavale.RealNodeValue, "-")
		if data[1] == "master" {
			SetDNDynamicMetanodeAddr(data[0])
			syncRun.Do(func() {
				startflag <- true
				close(startflag)
			})
		}
	}

	//定义打印计数器
	printCount := 0
	for {
		sort.Stable(WrapMetaValues(results))
		//以第一次master返回的数据作为种子节点
		rss, _ := json.MarshalIndent(results, "", "\t")

		if printCount > 20 {
			fmt.Println("更新时间:", time.Now().Format("2006-01-02 15:04:05"), ",metaNodes数据::", string(rss))
			printCount = 0
		}

		//查询出来的数据
		srearchdata := make([]MetaValue, 0)

	gotit:
		for _, metavale := range results {

			//查询提供元数据服务的信息
			cmd := &Cmd{Name: "get", Key: "metaNodes", Value: "metaNodes"}
			data := strings.Split(metavale.RealNodeValue, "-")
			livenodeAddr := data[0]
			client = NewTCPClient(livenodeAddr)
			//如果创建失败，说明该节点响应不了
			if client == nil {
				continue
			}
			//获取元服务节点
			cmd.Run(client)
			//如果出错也跳过当前遍历的节点，直接到下一个节点
			if cmd.Error != nil {
				continue
			} else {
				var rss []MetaValue
				json.Unmarshal([]byte(cmd.Value), &rss)
				for _, metavale := range rss {
					//当key为metaNodes，RealNodeValue值为:"nodeaddr-(master|slave)"
					data := strings.Split(metavale.RealNodeValue, "-")
					if data[1] == "master" {
						SetDNDynamicMetanodeAddr(data[0])
						syncRun.Do(func() {
							startflag <- true
							close(startflag)
						})
						//如果是master的话直接赋值给results，让其遍历最新获取的数据
						srearchdata = rss
						break gotit
					}
				}

				//如果返回的数据为空，说明当前节点还没有选择出master(处于不一致状态)，可以暂停100ms，重新让下一个节点取获取
				if GetDNDynamicMetanodeAddr() == "" {
					time.Sleep(100 * time.Millisecond)
					continue
				}

			}
		}
		//假如查询出来数据，赋值
		if len(srearchdata) > 0 {
			results = srearchdata
		} else {
			log.Warn("没有更新到dynamicMetanodeAddr.")
		}
		//间隔500毫秒刷新一次dynamicMetanodeAddr
		time.Sleep(500 * time.Millisecond)
		printCount++
	}
}
