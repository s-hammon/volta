package hl7

import "sync"

type msgData struct {
	Data map[string]interface{}
}

var pool = sync.Pool{
	New: func() interface{} {
		return &msgData{Data: make(map[string]interface{})}
	},
}

func getMsgMap() *msgData {
	return pool.Get().(*msgData)
}

func putMsgMap(m *msgData) {
	m.Data = make(map[string]interface{})
	pool.Put(m)
}
