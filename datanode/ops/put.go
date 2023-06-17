package ops

import (
	. "com.mgface.disobj/common"
	. "com.mgface.disobj/datanode/api"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func Put(writer http.ResponseWriter, req *http.Request) {
	url := req.URL.EscapedPath()
	//获取对象名称
	objName := strings.Split(url, "/")[2]
	//修改底层数据名称
	objRealName := objName + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	//例如："10.1.2.207:5000/sdp/20210207-10.1.2.207-5000/abc17787-1612644719934396484"
	filename := GetRDLocalStorePath() + string(os.PathSeparator) + objRealName
	realURL := ""
	if runtime.GOOS == "windows" {
		realURL = GetNodeAddr() + string(os.PathSeparator) + filename
	} else {
		realURL = GetNodeAddr() + filename
	}

	//对传输的数据进行hash运算
	var sourcedata []byte
	if req.Body != nil {
		sourcedata, _ = ioutil.ReadAll(req.Body)
	}
	hash := sha256.New()
	hash.Write(sourcedata)
	hashInBytes := hash.Sum(nil)
	sharedHashValue := hex.EncodeToString(hashInBytes)

	file, err := os.Create(filename)
	if err != nil {
		log.Debug(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	defer file.Close()
	file.Write(sourcedata)
	//强制把缓存页刷到硬盘
	file.Sync()

	index, _ := strconv.Atoi(req.Header.Get("index"))
	hashValue := req.Header.Get("hash")
	//记录文件的CRC
	status, e := bulidFileCRC(hashValue, sharedHashValue, realURL, index)
	writer.WriteHeader(status)
	if e != nil {
		writer.Write([]byte(e.Error()))
	}
}

//hashValue 完整文件的hash值
//
//sharedValue 分片文件的hash值
//
//realURL 分片存放路径
//
//index 分片索引号
func bulidFileCRC(hashValue, sharedHashValue, realURL string, index int) (int, error) {
	//做文件去重插入数据
	crc := &FileCRC{
		Data: make(map[string][]SharedData),
	}

	sharedData := SharedData{
		SharedFileUrlLocate: realURL,
		SharedFileHash:      sharedHashValue,
		SharedIndex:         index,
	}
	sd := make([]SharedData, 0)
	sd = append(sd, sharedData)
	crc.Data[hashValue] = sd
	data, _ := json.Marshal(crc)
	cmd := &Cmd{Name: "set", Key: "filecrc", Value: string(data)}

	client := NewReCallFuncTCPClient(GetDNDynamicMetanodeAddr, 3)
	if client == nil {
		tips := fmt.Sprintf("获取动态metanode值失败.")
		log.Warn(tips)
		return http.StatusInternalServerError, errors.New(tips)
	}

	cmd.Run(client)
	if cmd.Error != nil {
		return http.StatusInternalServerError, cmd.Error
	}
	return http.StatusOK, nil
}
