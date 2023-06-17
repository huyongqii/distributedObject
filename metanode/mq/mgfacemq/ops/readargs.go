package ops

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func readString(reader *bufio.Reader) (string, error) {
	tmp, e := reader.ReadString(' ')
	if e != nil {
		return "", e
	}
	key := strings.TrimSpace(tmp)
	return key, nil
}

func readLen(reader *bufio.Reader) (int, error) {
	tmp, e := reader.ReadString(' ')
	if e != nil {
		return 0, e
	}
	keylen, e := strconv.Atoi(strings.TrimSpace(tmp))
	if e != nil {
		return 0, e
	}
	return keylen, nil
}

//op<key.len>空格<value.len>空格<key><value>
func readKeyAndValue(reader *bufio.Reader) (string, []byte, error) {
	klen, e := readLen(reader) //读入key长度
	if e != nil {
		return "", nil, e
	}
	vlen, e := readLen(reader) //读入value长度
	if e != nil {
		return "", nil, e
	}
	kval := make([]byte, klen) //读入key值
	_, e = io.ReadFull(reader, kval)
	if e != nil {
		return "", nil, e
	}
	key := string(kval)

	vaules := make([]byte, vlen)
	_, e = io.ReadFull(reader, vaules) //读入value值
	if e != nil {
		return "", nil, e
	}
	//log.Println(fmt.Sprintf("key长度:%d,value长度:%d,key值:%s,value值:%s", klen, vlen, key, vaules))
	return key, vaules, nil
}
