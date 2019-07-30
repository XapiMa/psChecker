// +build !linux

package pschecker

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func (m *Monitor) monitor() error {
	for {
		go m.check()
		time.Sleep(time.Duration(m.interval) * time.Second)
	}
	return nil
}
func (m *Monitor) check() {
	fmt.Println("Checking processes infomretions")
	targets := make([]Target, 0)
	pses, err := process.Processes()
	if err != nil {
		log.Printf("Error: %s", errors.Wrap(err, "getProcessesInfo: processes"))
	}
	wg := &sync.WaitGroup{}
	ch := make(chan Target, 100)
	checkTypes := execFlag | cmdFlag | userFlag | pidFlag | openFlag
	for _, ps := range pses {
		wg.Add(1)
		go getProcessInfo(checkTypes, ps, wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	wgb := &sync.WaitGroup{}
	for target := range ch {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
		wgb.Add(1)
		go m.checkBlack(target, wgb)
		targets = append(targets, target)
	}
	wgw := &sync.WaitGroup{}
	wgw.Add(1)
	go m.checkWhiteAll(targets, wgw)
	wgw.Wait()
	wgb.Wait()
}

func (m *Monitor) initialWatcher() error {
	return nil
}
