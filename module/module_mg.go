package module

import (
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"sync"
)

type Module interface {
	OnInit()
	OnDestroy()
	Run(closeSig chan bool)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
}

var mods []*module

func Register(mi Module, name string) {
	m := new(module)
	m.mi = mi
	m.closeSig = make(chan bool, 1)

	// todo
	// using map to map name to module, map address to module.
	// test if the name conflict with each other.
	// give a new name if name == ""
	// search module for cluster.
	// explore API for cluster call.

	m.mi.OnInit()
	m.wg.Add(1)
	go run(m)

	mods = append(mods, m)
}

// func Init() {
// 	for i := 0; i < len(mods); i++ {
// 		mods[i].mi.OnInit()
// 	}

// 	for i := 0; i < len(mods); i++ {
// 		m := mods[i]
// 		m.wg.Add(1)
// 		go run(m)
// 	}
// }

func Destroy() {
	for i := len(mods) - 1; i >= 0; i-- {
		m := mods[i]
		m.closeSig <- true
		// wait the run to return. Reason is that, if the destory release the
		// resource which the goroutine in run is using, the program may panic.
		m.wg.Wait()
		destroy(m)
	}
}

func run(m *module) {
	m.mi.Run(m.closeSig)
	m.wg.Done()
}

func destroy(m *module) {
	// onDestory is defined by outside, means some kinds of danger of panic,
	// so recover if panic in OnDestroy.
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()

	m.mi.OnDestroy()
}
