package memory

import "encoding/json"

//返回自身的数据
func (cache *MemoryStore) ThisToJson() string {
	cache.Mutex.Lock()
	defer cache.Mutex.Unlock()
	//只传输这些心跳数据给其他mtanode数据
	dataTransfer := make(map[string]interface{}, 3)
	if v, ok := cache.Datas["dataNodes"]; ok {
		dataTransfer["dataNodes"] = v
	}
	if v, ok := cache.Datas["apiNodes"]; ok {
		dataTransfer["apiNodes"] = v
	}
	if v, ok := cache.Datas["metaNodes"]; ok {
		dataTransfer["metaNodes"] = v
	}
	data, _ := json.Marshal(dataTransfer)
	return string(data)
}
