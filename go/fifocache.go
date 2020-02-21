package main

import (
	"container/list"
	"sync"
)

var (
	fifomap  = make(map[string]struct{})
	fifolist = list.New()
	fifomu   = &sync.RWMutex{}
)

func cache_set(url string) bool {
	fifomu.Lock()
	defer fifomu.Unlock()

	_, ok := fifomap[url]
	if ok {
		return false
	}

	fifomap[url] = struct{}{}
	fifolist.PushFront(url)

	for fifolist.Len() > 10 {
		item := fifolist.Back()
		delete(fifomap, item.Value.(string))
		fifolist.Remove(item)
	}
	return true
}
