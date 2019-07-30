package pschecker

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

const (
	CACHE_ADD = iota
	CACHE_REMOVE
)

func (m *Monitor) monitor() error {
	return m.check()
}

func (m *Monitor) initialWatcher() error {
	var err error
	m.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "in initialWatcher")
	}
	if err = m.addProc(); err != nil {
		return errors.Wrap(err, "in initialWatcher")
	}
	return nil
}

func (m *Monitor) addProc() error {
	proc := "/proc"
	m.watcher.Add(proc)
	fileinfos, err := ioutil.ReadDir(proc)
	if err != nil {
		return errors.Wrap(err, "in addProc")
	}
	wg := &sync.WaitGroup{}
	for _, fi := range fileinfos {
		wg.Add(1)
		go m.addFd(fi.Name(), wg)
	}
	wg.Wait()
	return nil
}

func (m *Monitor) addFd(filename string, wgp *sync.WaitGroup) {
	if _, err := strconv.Atoi(filename); err == nil {
		m.watcher.Add(filepath.Join("/proc", filename))
	}
}

func (m *Monitor) check() error {
	fmt.Println("Checking processes infomretions")
	go m.firstCheck()

	for {
		select {
		case event := <-m.watcher.Events:
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				go m.pushProcess(event.Name)
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go m.popProcess(event.Name)
			}
		case err := <-m.watcher.Errors:
			return errors.Wrap(err, "in check")
		}
	}
}

func (m *Monitor) pushProcess(eventName string) {
	pid, err := strconv.Atoi(filepath.Base(eventName))
	if err != nil {
		return
	}
	if _, ok := m.cache[pid]; ok {
		return
	}
	m.cache[pid], err = getProc(pid)
	if err != nil {
		log.Print(err)
	}
	m.checkList(m.cache[pid], CACHE_ADD)
}

func (m *Monitor) popProcess(eventName string) {
	pid, err := strconv.Atoi(filepath.Base(eventName))
	if err != nil {
		return
	}
	if _, ok := m.cache[pid]; !ok {
		return
	}
	m.checkList(m.cache[pid], CACHE_REMOVE)
	delete(m.cache, pid)
}

func (m *Monitor) firstCheck() {

}

func (m *Monitor) checkList(p Proc, flag int) {
	switch flag {
	case CACHE_ADD:
	case CACHE_REMOVE:
	}
}
