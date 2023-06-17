package memory

import (
	. "com.mgface.disobj/common"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"reflect"
)

func (cache *MemoryStore) Get(key string, value []byte) (interface{}, error) {
	cache.Mutex.RLock()
	defer cache.Mutex.RUnlock()
	data := make([]interface{}, 0)
	if key == "dataNodes" || key == "apiNodes" || key == "metaNodes" {
		mvalues, exit := cache.Datas[key]
		if exit {
			log.Debug("操作:", key, ",value:", string(value), ",当前类型：", reflect.TypeOf(mvalues))

			var forceTxData []MetaValue
			datas, _ := json.Marshal(mvalues)
			json.Unmarshal(datas, &forceTxData)

			for _, val := range forceTxData {
				data = append(data, val)
			}
			return data, nil
		}
	} else if key == "metadata" {
		sysvalues, exit := cache.Datas[key]
		if exit {
			log.Debug("操作:", key, ",value:", string(value), ",当前类型：", reflect.TypeOf(sysvalues))

			var forceTxData map[string][]Datadigest
			datas, _ := json.Marshal(sysvalues)
			json.Unmarshal(datas, &forceTxData)
			if v, ok := forceTxData[string(value)]; ok {
				return v, nil
			}
		}
	} else if key == "filecrc" {
		sysvalues, exit := cache.Datas[key]
		if exit {
			log.Debug("操作:", key, ",value:", string(value), ",当前类型：", reflect.TypeOf(sysvalues))

			var forceTxData map[string][]SharedData
			datas, _ := json.Marshal(sysvalues)
			json.Unmarshal(datas, &forceTxData)

			if v, ok := forceTxData[string(value)]; ok {
				return v, nil
			}
		}
	} else {
		return cache.Datas[key], nil
	}
	return data, nil
}
