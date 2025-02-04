package hl7

import "sync"

var pool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

func getMsgMap() map[string]interface{} {
	return pool.Get().(map[string]interface{})
}

// func putMsgMap(m map[string]interface{}) {
// 	for k := range m {
// 		delete(m, k)
// 	}
// 	pool.Put(m)
// }
