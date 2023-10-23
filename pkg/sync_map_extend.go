package pkg

import "sync"

// @Author :George
// @File: sync_map_extend
// @Version: 1.0.0
// @Date 2023/10/10 17:01

type MySyncMap struct {
	sync.Map
}

func (m *MySyncMap) Len() int {
	leng := 0
	m.Range(func(k, v any) bool {
		leng++
		return true
	})
	return leng
}
