package pschecker

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

const (
	cacheAdd = iota
	cacheRemove
	cacheChange
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
		go m.addFd(filepath.Join(proc, fi.Name()), wg)
	}
	wg.Wait()
	return nil
}

func (m *Monitor) addFd(path string, wgp *sync.WaitGroup) {
	if isProcDir(path) {
		m.watcher.Add(filepath.Join(path, "fd"))
	}
}

func isProcDir(path string) bool {
	path = filepath.Clean(path)
	dir, base := filepath.Split(path)

	if filePath.Clean(dir) != "/proc" {
		return false
	}
	return isNum(base)
}
func isNum(s string) bool {
	if _, err := strconv.Atoi(s); err != nil {
		return false
	}
	return true
}

func isFdDir(path string) bool {
	path = filepath.Clean(path)
	p := strings.Split(path, "/")
	if p[1] == "proc" && isNum(p[2]) && p[3] == "fd" {
		return true
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
				if isProcDir(event.Name) {
					go m.pushProc(event.Name)
				} else if isfdDir(event.Name) {
					go m.changeProc(openFlag, event.Name)
				}
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				if isProcDir(event.Name) {
					go m.popProc(event.Name)
				} else if isfdDir(event.Name) {
					go m.changeProc(openFlag, event.Name)
				}
			}
		case err := <-m.watcher.Errors:
			return errors.Wrap(err, "in check")
		}
	}
}

func (m *Monitor) pushProc(path string) {
	pid, err := getPidFromProcPath(path)
	if err != nil {
		log.Print(err)
		return
	}
	if _, ok := m.cache[pid]; ok {
		return
	}
	proc, err := getProc(pid)
	if err != nil {
		log.Print(err)
	}
	m.cache[pid] = proc
	m.checkList(proc, cacheAdd, userFlag|pidFlag|openFlag|cmdFlag|execFlag)
}

func (m *Monitor) popProc(path string) {
	pid, err := getPidFromProcPath(path)
	if err != nil {
		log.Print(err)
		return
	}
	if _, ok := m.cache[pid]; !ok {
		return
	}
	m.checkList(m.cache[pid], cacheRemove, userFlag|pidFlag|openFlag|cmdFlag|execFlag)
	delete(m.cache, pid)
}

func (m *Monitor) changeProc(path string, flag int) {
	pid, err := getPidFromProcPath(path)
	if err != nil {
		log.Print(err)
		return
	}
	if _, ok := m.cache[pid]; !ok {
		return
	}
	wg := &sync.WaitGroup{}

	if flag&openFlag != 0 {
		ps := process.NewProcess(int32(pid))
		wg.Add(1)
		go m.cache[pid].setOpen(ps, wg)
	}

	wg.Wait()
	m.checkList(m.cache[pid], cacheChange, flag)
}

func getPidFromProcPath(path string) (int, error) {
	path = filepath.Clean(path)
	p := strings.Split(path, "/")
	if len(p) < 3 {
		return 0, fmt.Errorf("%s is not proc path", path)
	}
	pid, err := strconv.Atoi(path[2])
	if err != nil {
		return 0, fmt.Errorf("%s is not proc path", path)
	}
	return pid, nil
}

func (m *Monitor) checkList(p Proc, actionFlag int, itemFlag int) {

	wg := &sync.WaitGroup{}

	switch actionFlag {
	case cacheAdd:
		wg.Add(1)
		go m.inNotWhite(p, itemFlag, wg)
		wg.Add(1)
		go m.inBlack(p, itemFlag, wg)
	case cacheRemove:
		wg.Add(1)
		go m.ExistWhite(p, itemFlag, wg)
		wg.Add(1)
		go m.ExistBlack(p, itemFlag, wg)
	case cacheChange:
		wg.Add(1)
		go m.inNotWhite(p, itemFlag, wg)
	}
}

func (m *Monitor) firstCheck() {

}

func (m *Monitor) addChecklist(p Proc, pwg *sync.WaitGroup) {
	defer pwg.Done()
	if m.whitelist.in(p) {
		m.wc.add(p)
	}
	if m.blacklist.in(p) {
		m.bc.add(p)
	}
}

// → whitelistに含まれるか
// 	含まれる → whitecacheに追加・通知
// 	含まれない → そのまま
// → blacklistに含まれるか
// 	含まれる → blackcacheに追加・通知
// 	含まない → そのまま

// 削除される
// → whitecacheに含まれるか
// 	含まれる → whitecacheから削除・通知
// 		まだ該当するwhitelistを満たす要素はまだあるか
// 			ない → 通知
// 			ある → そのまま
// 	含まれない → そのまま
// → blackcacheに含まれるか
// 	含まれる → blackcacheから削除・通知
// 		該当するblacklistを満たす要素はまだあるか
// 			ない → 通知
// 			ある → そのまま
// 	含まれない → そのまま

// 変更される
// → whitecacheに含まれるか
// 	含まれる → まだ該当するwhitelistに合致するか
// 		合致する → 更新
// 		合致しない → 削除・通知
// 			まだ該当するwhitelistに合致するプロセスが存在するか
// 				存在しない → 通知
// 				存在する → そのまま
// 	含まれない → 該当するno whitelist cacheが存在するか
// 		存在する → 追加・通知
// 		存在しない → そのまま
// → blackcacheに含まれるか
// 	含まれる → まだ該当するblacklistに合致するか
// 		合致する → 更新
// 		合致しない → 削除・通知
// 			まだ該当のblacklistに合致するプロセスが存在するか
// 				存在しない → 通知
// 				存在する → そのまま
// 	含まれない → 該当する no blacklist cacheが存在するか
// 		存在する → 追加・通知
// 		存在しない → そのまま
