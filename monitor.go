package pschecker

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// Monitor is struct of m command
type Monitor struct {
	outputPath string
	whitelist  []Target
	blacklist  []Target
	interval   int
	regexp     map[string]*regexp.Regexp
	watcher    *fsnotify.Watcher
	cache      map[int]Proc
}

// Target is item struct of whitelist and blacklist
type Target struct {
	Exec   string   `yaml:"exec"`
	Cmd    string   `yaml:"cmd"`
	Open   []string `yaml:"open"`
	User   string   `yaml:"user"`
	Pid    int      `yaml:"pid"`
	Regexp string   `yaml:"regexp"`
}

// Proc is struct of fond process data
type Proc struct {
	Exec string
	Cmd  string
	Open []string
	User string
	Pid  int
}

// NewMonitor create new m object
func NewMonitor(whitelistPath, blacklistPath, outputPath string, interval int) (*Monitor, error) {
	m := new(Monitor)
	var err error
	m.whitelist, err = readList(whitelistPath)
	if err != nil {
		return m, errors.Wrap(err, "in NewMonitor when read whitelist")
	}
	m.blacklist, err = readList(blacklistPath)
	if err != nil {
		return m, errors.Wrap(err, "in NewMonitor when read blacklist")
	}
	m.regexp = make(map[string]*regexp.Regexp)
	if err := m.addRegexp(m.whitelist); err != nil {
		return m, errors.Wrap(err, "in NewMonitor when add regexp item of whitelist")
	}
	if err := m.addRegexp(m.blacklist); err != nil {
		return m, errors.Wrap(err, "in NewMonitor when add regexp item of blaclist")
	}
	m.interval = interval
	m.outputPath = outputPath
	m.cache = make(map[int]Proc)
	if err := m.initialWatcher(); err != nil {
		return m, err
	}
	return m, nil
}

func readList(listPath string) ([]Target, error) {
	list, err := parseConfigYml(listPath)
	if err != nil {
		return list, errors.Wrapf(err, "cause in NewMonitor: for %s", listPath)
	}
	for _, target := range list {
		sort.Slice(target.Open, func(i, j int) bool { return target.Open[i] < target.Open[j] })
	}
	return list, nil
}

func (m *Monitor) addRegexp(list []Target) error {
	if m.regexp == nil {
		m.regexp = make(map[string]*regexp.Regexp)
	}
	for i, target := range list {
		if _, ok := m.regexp[target.Regexp]; !ok {
			r, err := regexp.Compile(target.Regexp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("in %dth item's regexp string %q can't compile", i+1, target.Regexp))
			}
			m.regexp[target.Regexp] = r
		}
	}
	return nil
}

// Monitor is processes monitoring function
func (m *Monitor) Monitor() error {
	return m.monitor()
}

type foundTarget struct {
	found  bool
	target Target
}

func (m *Monitor) checkWhiteAll(targets []Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")

	wg := &sync.WaitGroup{}
	ch := make(chan foundTarget, 100)
	for _, white := range m.whitelist {
		wg.Add(1)
		go m.isWhite(white, targets, wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		if !result.found {
			outputTxt := fmt.Sprintf("%s NotFound: %v\n", timeString, result.target)
			if err := appendFile(m.outputPath, outputTxt); err != nil {
				log.Print(errors.Wrap(err, "couse in checkWhite"))
			}
		}
	}
}

func (m *Monitor) checkBlack(target Target, wgp *sync.WaitGroup) {
	defer wgp.Done()
	timeString := time.Now().Format("2006/01/02 15:04:05")
	found := m.isBlack(target)
	if found {
		outputTxt := fmt.Sprintf("%s WARNING!! FOUND!!: %v\n", timeString, target)
		if err := appendFile(m.outputPath, outputTxt); err != nil {
			log.Print(errors.Wrap(err, "couse in checkBlack"))
		}
	}
}

func (m *Monitor) isWhite(white Target, targets []Target, wgp *sync.WaitGroup, chp chan foundTarget) {
	defer wgp.Done()

	wg := &sync.WaitGroup{}
	ch := make(chan bool)
	for _, target := range targets {
		wg.Add(1)
		go m.isWhiteFunc(white, target, wg, ch)
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

func (m *Monitor) isWhiteFunc(white Target, target Target, wgp *sync.WaitGroup, ch chan bool) {
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
			if ok, err := m.matchPattern(target.Exec, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			if ok, err := m.matchPattern(target.Cmd, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			if ok, err := m.matchPattern(target.User, white.Regexp); err != nil {
				log.Print(errors.Wrap(err, "couse in isWhite"))
			} else if ok {
				reFlag = true
				break
			}
			for _, open := range target.Open {
				if ok, err := m.matchPattern(open, white.Regexp); err != nil {
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
func (m *Monitor) isBlack(target Target) bool {
	found := false
	for _, black := range m.blacklist {
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
				if ok, err := m.matchPattern(target.Exec, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				if ok, err := m.matchPattern(target.Cmd, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				if ok, err := m.matchPattern(target.User, black.Regexp); err != nil {
					log.Print(errors.Wrap(err, "couse in isBlack"))
				} else if ok {
					reFlag = true
					break
				}
				for _, open := range target.Open {
					if ok, err := m.matchPattern(open, black.Regexp); err != nil {
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

func (m *Monitor) matchPattern(str, pattern string) (bool, error) {
	re, ok := m.regexp[pattern]
	if !ok {
		return false, fmt.Errorf("not found: regexp of %s", pattern)
	}
	return matchPattern(str, re), nil
}

func matchPattern(str string, r *regexp.Regexp) bool {
	return r.Match([]byte(str))
}
