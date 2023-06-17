package ops

import (
	"bufio"
	. "com.mgface.disobj/metanode/store"
)

//abnf规则]G<key.len> <key>
func Get(store Store, reader *bufio.Reader) (interface{}, error) {
	key, value, _ := readKeyAndValue(reader)
	return store.Get(key, value)
}

func Set(store Store, reader *bufio.Reader) error {
	key, value, e := readKeyAndValue(reader)
	store.Set(key, value)
	return e
}

func Del(store Store, reader *bufio.Reader) error {
	key, value, e := readKeyAndValue(reader)
	store.Del(key, value)
	return e
}

//文件上传保持接口
func Put(store Store, reader *bufio.Reader) (interface{}, error) {
	fname, _ := readString(reader)

	fsize, _ := readLen(reader)
	return store.Put(fname, fsize, reader)
}

func Syn(store Store, reader *bufio.Reader) (interface{}, error) {
	synrlen, _ := readLen(reader)
	return store.Sync(synrlen, reader)
}
