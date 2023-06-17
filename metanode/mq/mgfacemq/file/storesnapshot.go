package file

import (
	. "com.mgface.disobj/common"
	. "com.mgface.disobj/metanode/mq/mgfacemq/memory"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sort"
	"sync/atomic"
	"time"
)

//存储内存数据快照到文件系统中
func StoresnapshotData(ms *MemoryStore) {
	storepath := ms.StorePath

	//清理log文件
	//go clearLogfile(storepath)

again:
	dateflag := time.Now().Format("20060102")
	filename := fmt.Sprintf("%s%v%v%s", storepath, string(os.PathSeparator), dateflag, ".log")

	//获取当前时间，放到now里面，要给next用
	now := time.Now()
	//通过now偏移24小时
	next := now.Add(time.Hour * 24)
	//获取下一个凌晨的日期
	next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	//计算当前时间到凌晨的时间间隔，设置一个定时器
	t := time.NewTimer(next.Sub(now))

	//buf := bufio.NewWriter(file)
	flushTicker := time.NewTicker(5 * time.Second)
	data := make([]RecMsg, 0)
	var activeBuff int32
	for {
		select {
		case msg := <-ms.Msgs:
			data = append(data, msg)
			if atomic.LoadInt32(&activeBuff) == 1 {
				//同步
				ms.BuffMsgs <- msg
			}
		case <- ms.BuffSemaphore: //如果是同步数据
			log.Info("当前还剩未处理size:", len(data), ",sync时间:", time.Now().Format("2006-01-02 15:04:05"))
			file, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
			//如果存在数据，那么还是回写到前一天的日志文件中
			if len(data) > 0 {
				datax, _ := json.Marshal(data)
				batchsize := len(datax)
				batchdata := make([]byte, 0)
				batchdata = append(batchdata, IntToBytes(batchsize)...)
				//后写消息字节数据
				batchdata = append(batchdata, datax...)
				file.Write(batchdata)
				file.Sync()

			}
			//从头开始读取该文件
			file.Seek(0, io.SeekStart)
			//复制文件
			syncFd := fmt.Sprintf("%s.%s", filename, "sync")
			syncfile, _ := os.OpenFile(syncFd, os.O_WRONLY|os.O_CREATE, 0644)
			buf := make([]byte, 4096)
			for {
				n, e := file.Read(buf)
				if e == io.EOF {
					break
				}
				syncfile.Write(buf[:n])
				syncfile.Sync()
			}

			file.Close()
			syncfile.Close()
			//todo 快照文件和缓冲可以同时操作，不要互相影响
			ms.FishedSnapshot <- true
			//激活sync消息通道
			atomic.StoreInt32(&activeBuff, 1)

		case <-flushTicker.C:
			if len(data) > 0 {
				//算出该数据的大小
				datax, _ := json.Marshal(data)
				batchsize := len(datax)
				batchdata := make([]byte, 0)
				//先写消息长度,长度为4的字节
				batchdata = append(batchdata, IntToBytes(batchsize)...)
				//后写消息字节数据
				batchdata = append(batchdata, datax...)

				file, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
				//file.Seek(0, io.SeekEnd)

				file.Write(batchdata)
				file.Sync()
				file.Close()
				//对data进行重置
				data = make([]RecMsg, 0)
			}
		case <-t.C: //已经到了凌晨第二天了
			log.Info("当前还剩未处理size:", len(data), ",已经翻滚到第二天日期:", time.Now().Format("2006-01-02 15:04:05"))
			//如果存在数据，那么还是回写到前一天的日志文件中
			if len(data) > 0 {
				datax, _ := json.Marshal(data)
				batchsize := len(datax)
				batchdata := make([]byte, 0)
				batchdata = append(batchdata, IntToBytes(batchsize)...)
				batchdata = append(batchdata, datax...)

				file, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)

				file.Write(batchdata)
				file.Sync()
				//设置文件为只读模式
				os.Chmod(filename, 0444)
				file.Close()
			}
			goto again
		}
	}
}

//清理元数据log文件，只保留最新的五个文件
func clearLogfile(storepath string) {
	for {
		time.Sleep(time.Duration(3600) * time.Second)
		files := WalkDirectory(storepath)
		//移除老的文件
		if len(files) > 5 {
			tfs := FileDescs(files)
			sort.Stable(tfs)
			tfs = tfs[5:]
			for _, v := range tfs {
				os.Remove(v.Fpath)
			}
		}
	}

}
