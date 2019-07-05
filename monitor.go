package pschecker

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

// Monitor is struct of monitor command
type Monitor struct {
	outputPath string
	whitelist  []Target
	blacklist  []Target
	interval   int
	regexp     map[string]*regexp.Regexp
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
	m.regexp = make(map[string]*regexp.Regexp)
	for i, target := range m.whitelist {
		if _, ok := m.regexp[target.Regexp]; !ok {
			r, err := regexp.Compile(target.Regexp)
			if err != nil {
				return m, errors.Wrap(err, fmt.Sprintf("in whitelist %dth item's regexp string %q can't compile", i+1, target.Regexp))
			}
			m.regexp[target.Regexp] = r
		}

	}
	m.blacklist, err = parseConfigYml(blacklistPath)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor: for blacklist")
	}
	for _, target := range m.blacklist {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
	}
	for i, target := range m.blacklist {
		if _, ok := m.regexp[target.Regexp]; !ok {
			r, err := regexp.Compile(target.Regexp)
			if err != nil {
				return m, errors.Wrap(err, fmt.Sprintf("in blacklist %dth item's regexp string %q can't compile", i+1, target.Regexp))
			}
			m.regexp[target.Regexp] = r
		}
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
		go monitor.checkBlack(target, wgb)
		targets = append(targets, target)
	}
	wgw := &sync.WaitGroup{}
	wgw.Add(1)
	go monitor.checkWhiteAll(targets, wgw)
	wgw.Wait()
	wgb.Wait()
}

type foundTarget struct {
	found  bool
	target Target
}

func (monitor *Monitor) checkWhiteAll(targets []Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")

	wg := &sync.WaitGroup{}
	ch := make(chan foundTarget, 100)
	for _, white := range monitor.whitelist {
		wg.Add(1)
		go monitor.isWhite(white, targets, wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		if !result.found {
			outputTxt := fmt.Sprintf("%s NotFound: %v\n", timeString, result.target)
			if err := appendFile(monitor.outputPath, outputTxt); err != nil {
				log.Print(errors.Wrap(err, "couse in checkWhite"))
			}
		}
	}
}

func (monitor *Monitor) checkBlack(target Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")
	found := monitor.isBlack(target)
	if found {
		outputTxt := fmt.Sprintf("%s WARNING!! FOUND!!: %v\n", timeString, target)
		if err := appendFile(monitor.outputPath, outputTxt); err != nil {
			log.Print(errors.Wrap(err, "couse in checkBlack"))
		}
	}
}

func (monitor *Monitor) isWhite(white Target, targets []Target, wgp *sync.WaitGroup, chp chan foundTarget) {
	defer wgp.Done()

	wg := &sync.WaitGroup{}
	ch := make(chan bool)
	for _, target := range targets {
		wg.Add(1)
		go monitor.isWhiteFunc(white, target, wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var found bool
	for found = range ch {
		if found {
			break
		}
	}

	result := foundTarget{found, white}
	chp <- result
}

func (monitor *Monitor) isWhiteFunc(white Target, target Target, wgp *sync.WaitGroup, ch chan bool) {
	defer wgp.Done()

	found := true
	if white.Exec != "" && white.Exec != target.Exec {
		found = false
	}
	if white.Cmd != "" && white.Cmd != target.Cmd {
		found = false
	}
	if len(white.Open) != 0 {
		if fmt.Sprintf("%v", white.Open) != fmt.Sprintf("%v", target.Open) {
			found = false
		}
	}
	if white.User != "" && white.User != target.User {
		found = false
	}
	if white.Pid != 0 && white.Pid != target.Pid {
		found = false
	}
	if white.Regexp != "" {
		reFlag := false
		for x := 0; x < 1; x++ {
			if ok, err := monitor.matchPattern(target.Exec, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			if ok, err := monitor.matchPattern(target.Cmd, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			if ok, err := monitor.matchPattern(target.User, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			for _, open := range target.Open {
				if ok, err := monitor.matchPattern(open, white.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isWhite"))
				} else if ok {
					reFlag = true
					break
				}
			}
		}
		if !reFlag {
			found = false
		}
	}
	ch <- found

}
func (monitor *Monitor) isBlack(target Target) bool {
	found := false
	for _, black := range monitor.blacklist {
		found = true
		if black.Exec != "" && target.Exec != black.Exec {
			found = false
		}
		if black.Cmd != "" && target.Cmd != black.Cmd {
			found = false
		}
		if len(black.Open) != 0 {
			if fmt.Sprintf("%v", target.Open) != fmt.Sprintf("%v", black.Open) {
				found = false
			}
		}
		if black.User != "" && target.User != black.User {
			found = false
		}
		if black.Pid != 0 && target.Pid != black.Pid {
			found = false
		}
		if black.Regexp != "" {
			reFlag := false
			for x := 0; x < 1; x++ {
				if ok, err := monitor.matchPattern(target.Exec, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				if ok, err := monitor.matchPattern(target.Cmd, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				if ok, err := monitor.matchPattern(target.User, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				for _, open := range target.Open {
					if ok, err := monitor.matchPattern(open, black.Regexp); err != nil {
						log.Print(errors.Wrap(err, "couse in isBlack"))
					} else if ok {
						reFlag = true
						break
					}
				}
			}
			if !reFlag {
				found = false
			}
		}
		if found {
			break
		}
	}
	return found
}

func (monitor *Monitor) matchPattern(str, pattern string) (bool, error) {
	re, ok := monitor.regexp[pattern]
	if !ok {
		return false, fmt.Errorf("not found: regexp of %s", pattern)
	}
	return matchPattern(str, re), nil
}

func matchPattern(str string, r *regexp.Regexp) bool {
	return r.Match([]byte(str))
}
