package api

import (
	. "com.mgface.disobj/common"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
	"strings"
)

func Locate(object string, maxFailCount int) (nodeAddrs []string, objRealNames []string, filesize int64) {
	failCount := 0
	for {
		client := NewTCPClient(GetDynamicMetanodeAddr())
		if client == nil {
			log.Warn("元数据节点连接失败，正在重试中......")
			failCount++
			if failCount > maxFailCount {
				log.Warn(fmt.Sprintf("元数据节点重试%d次失败，定位失败.", maxFailCount+1))
				return
			}
			continue
		}
		req := NewRequest(client, "get", "metadata", object)
		err := req.Run()
		if err != nil {
			return nil, nil, 0
		}

		var results []Datadigest
		err = json.Unmarshal([]byte(req.GetValue()), &results)
		if err != nil {
			return nil, nil, 0
		}

		// 获取最新的版本
		var max int64 = 0
		var maxDataDigest Datadigest
		for _, data := range results {
			if data.Version > max {
				max = data.Version
				maxDataDigest = data
			}
		}

		// 通过hash值获取数据
		req = NewRequest(client, "get", "filecrc", maxDataDigest.Hash)
		err = req.Run()
		if err != nil {
			return nil, nil, 0
		}
		var dataSlices SharedDataslice
		err = json.Unmarshal([]byte(req.GetValue()), &dataSlices)
		if err != nil {
			return nil, nil, 0
		}
		sort.Stable(dataSlices)

		nodeAddrs = make([]string, 0)
		objRealNames = make([]string, 0)

		for _, data := range dataSlices {
			url := strings.Split(data.SharedFileUrlLocate, string(os.PathSeparator))
			nodeAddrs = append(nodeAddrs, url[0])
			objRealName := fmt.Sprintf("%s/%s", url[len(url)-2], url[len(url)-1])
			log.Debug("objRealName:::", objRealName)
			objRealNames = append(objRealNames, objRealName)
		}

		return nodeAddrs, objRealNames, maxDataDigest.Datasize
	}
}

type SharedDataslice []SharedData

// Len 数据分片切片长度
func (s SharedDataslice) Len() int {
	return len(s)
}

// Swap 数据分片的数据交换
func (s SharedDataslice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less 数据分片的比较
func (s SharedDataslice) Less(i, j int) bool {
	return s[i].SharedIndex < s[j].SharedIndex
}
