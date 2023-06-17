package memory

import (
	. "com.mgface.disobj/common"
	"encoding/json"
	"time"
)

func (cache *MemoryStore) Set(key string, value []byte) error {
	//假如为非心跳数据，那么则发送消息给消息队列
	if !(key == "dataNodes" || key == "apiNodes" || key == "metaNodes") && !cache.LoadingSnapshot {
		//todo 传递string而不是[]byte，主要是json.Marshal对字节数组回进行base64，导致字符串足够长,不利于后续压缩算法
		cache.Msgs <- RecMsg{Key: key, Val: string(value)}
	}
	cache.Mutex.Lock()
	defer cache.Mutex.Unlock()
	sysvalues, exit := cache.Datas[key]
	if key == "dataNodes" || key == "apiNodes" || key == "metaNodes" {
		v := MetaValue{RealNodeValue: string(value), Created: time.Now()}
		data := make([]MetaValue, 0)
		data = append(data, v)
		if exit {
			var forceTxData []MetaValue
			datas, _ := json.Marshal(sysvalues)
			json.Unmarshal(datas, &forceTxData)
			for _, val := range forceTxData {
				//如果存在同样datanode IP的数据，那么移除老数据，添加新数据
				if val.RealNodeValue != v.RealNodeValue {
					data = append(data, val)
				}
			}
		}
		cache.Datas[key] = data
		return nil
	} else if key == "metadata" {

		//假如metadata不存在，那么进行初始化操作
		var valueDKV DataKeyValue
		json.Unmarshal(value, &valueDKV) //解析value的数据到dkv变量
		if exit {

			var loadsysvalues map[string][]Datadigest
			datas, _ := json.Marshal(sysvalues)
			json.Unmarshal(datas, &loadsysvalues)

			for k, v := range valueDKV.Data {
				if vdata, ok := loadsysvalues[k]; ok { //假如存在，那么数据添加

					var max int64 = 0
					for _, v := range vdata {
						if v.Version > max {
							max = v.Version
						}
					}
					//修改当前附加的版本号
					v[0].Version = max + 1
					loadsysvalues[k] = append(vdata, v...)
				} else {
					loadsysvalues[k] = v
				}
				cache.Datas[key] = loadsysvalues
				return nil
			}
		}
		//假如数据不存在
		cache.Datas[key] = valueDKV.Data
	} else if key == "filecrc" {
		var filecrc FileCRC
		json.Unmarshal(value, &filecrc)
		if exit {
			//加载内存数据
			var loadsysvalues map[string][]SharedData
			datas, _ := json.Marshal(sysvalues)
			json.Unmarshal(datas, &loadsysvalues)
			for k, v := range filecrc.Data {
				if vdata, ok := loadsysvalues[k]; ok { //假如存在，那么数据添加
					loadsysvalues[k] = append(vdata, v...)
				} else {
					loadsysvalues[k] = v
				}
			}
			cache.Datas[key] = loadsysvalues
			return nil
		}
		//假如数据不存在
		cache.Datas[key] = filecrc.Data
	} else {
		v := MetaValue{RealNodeValue: string(value), Created: time.Now()}
		cache.Datas[key] = v
	}
	return nil
}
