package file

import (
	. "com.mgface.disobj/common"
	. "com.mgface.disobj/metanode/mq/mgfacemq/memory"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

//加载内存数据快照到文件系统中
func LoadsnapshotData(cache *MemoryStore) {
	storepath := cache.StorePath
	//目录不存在直接跳过加载
	if _, err := os.Stat(storepath); os.IsNotExist(err) {
		return
	}
	fpaths := readLogfile(storepath)
	for _, fpath := range fpaths {
		//跳过同步给其他节点的文件
		if strings.Contains(fpath, ".sync") {
			continue
		}
		log.Info(fmt.Sprintf("加载快照文件:%s", fpath))
		f, _ := os.OpenFile(fpath, os.O_RDONLY, 0644)
		for {
			msgsize := make([]byte, 4)
			_, e := f.Read(msgsize)
			if e == io.EOF {
				break
			}
			size := BytesToInt(msgsize)
			ev := make([]byte, size)
			f.Read(ev)
			var msgs []RecMsg
			json.Unmarshal(ev, &msgs)
			for _, v := range msgs {
				cache.Set(v.Key, []byte(v.Val.(string)))
			}
		}
		f.Close()
	}
	cache.LoadingSnapshot = false //加载快照文件完成
}

func readLogfile(storepath string) []string {
	fpaths := make([]string, 0)

	fds := WalkDirectory(storepath)

	for _, v := range fds {
		fpaths = append(fpaths, v.Fpath)
	}
	return fpaths
}
