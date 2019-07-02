package pschecker

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Monitor is struct of monitor command
type Monitor struct {
	outputPath string
	whitelist  []Target
	blacklist  []Target
	interval   int
}

// NewMonitor create new monitor object
func NewMonitor(whitelistPath, blacklistPath, outputPath string, interval int) (*Monitor, error) {
	m := new(Monitor)
	var err error
	m.whitelist, err = parseConfigYml(whitelistPath)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor: for whitelist")
	}
	for _, target := range m.whitelist {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
	}
	m.blacklist, err = parseConfigYml(blacklistPath)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor: for blacklist")
	}
	for _, target := range m.blacklist {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
	}
	m.interval = interval
	m.outputPath = outputPath
	return m, nil
}

// Monitor is processes monitorring function
func (monitor *Monitor) Monitor() error {

	for {
		go monitor.psCheck()
		time.Sleep(time.Duration(monitor.interval) * time.Second)
	}

	return nil
}

func (monitor *Monitor) psCheck() {
	fmt.Println("Checking processes infomretions")
	targets, err := getProcessesInfo(execFlag | cmdFlag | userFlag | pidFlag | openFlag)
	if err != nil {
		log.Printf("Error: %s\n", errors.Wrap(err, "cause in psCheck"))
	}
	for _, target := range targets {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go monitor.checkWhite(targets, wg)
	wg.Add(1)
	go monitor.checkBlack(targets, wg)
	wg.Wait()
}

func (monitor *Monitor) checkWhite(targets []Target, wg *sync.WaitGroup) {
	defer wg.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")

	for _, white := range monitor.whitelist {
		found := monitor.isExistPs(white, targets)
		if !found {
			outputTxt := fmt.Sprintf("%s NotFound: %v\n", timeString, white)
			if err := appendFile(monitor.outputPath, outputTxt); err != nil {
				log.Print(errors.Wrap(err, "couse in checkWhite"))
			}
		}
	}
}

func (monitor *Monitor) checkBlack(targets []Target, wg *sync.WaitGroup) {
	defer wg.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")

	for _, target := range targets {
		found := monitor.isExistPsBlack(target, monitor.blacklist)
		if found {
			outputTxt := fmt.Sprintf("%s WARNING!! FOUND!!: %v\n", timeString, target)
			if err := appendFile(monitor.outputPath, outputTxt); err != nil {
				log.Print(errors.Wrap(err, "couse in checkBlack"))
			}
		}
	}
}

func (monitor *Monitor) isExistPs(item Target, targets []Target) bool {
	found := false
	for _, target := range targets {
		found = true
		if item.Exec != "" && item.Exec != target.Exec {
			found = false
		}
		if item.Cmd != "" && item.Cmd != target.Cmd {
			found = false
		}
		if len(item.Open) != 0 {
			if fmt.Sprintf("%v", item.Open) != fmt.Sprintf("%v", target.Open) {
				found = false
			}
		}
		if item.User != "" && item.User != target.User {
			found = false
		}
		if item.Pid != 0 && item.Pid != target.Pid {
			found = false
		}
		if found {
			break
		}
	}
	return found
}

func (monitor *Monitor) isExistPsBlack(item Target, targets []Target) bool {
	found := false
	for _, target := range targets {
		found = true
		if target.Exec != "" && item.Exec != target.Exec {
			found = false
		}
		if target.Cmd != "" && item.Cmd != target.Cmd {
			found = false
		}
		if len(item.Open) != 0 {
			if fmt.Sprintf("%v", item.Open) != fmt.Sprintf("%v", target.Open) {
				found = false
			}
		}
		if target.User != "" && item.User != target.User {
			found = false
		}
		if target.Pid != 0 && item.Pid != target.Pid {
			found = false
		}
		if found {
			break
		}
	}
	return found
}
