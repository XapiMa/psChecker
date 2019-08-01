package psmonitor

import (
	"fmt"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// Monitor is struct of monitor command
type Monitor struct {
	wcc        cacheCount
	bcc        cacheCount
	cache      map[int]cacheItem
	outputPath string
	regexp     map[string]*regexp.Regexp
	watcher    *fsnotify.Watcher
	message    chan message
}
type message interface {
}

// NewMonitor create new monitor object
func NewMonitor(whitelistPath, blacklistPath, outputPath string, interval int) (*Monitor, error) {
	// ToDo: delete the intervel variable

	var err error

	m := new(Monitor)

	m.wcc, err = readList(whitelistPath)
	if err != nil {
		return m, errors.Wrap(err, "in NewMonitor read wcc")
	}

	m.bcc, err = readList(blacklistPath)
	if err != nil {
		return m, errors.Wrap(err, "in NewMonitor read bcc")
	}

	m.cache = make(map[int]cacheItem)

	m.outputPath = outputPath

	m.regexp = make(map[string]*regexp.Regexp)
	if err := m.addRegexp(m.wcc); err != nil {
		return m, errors.Wrap(err, "in NewMonitor when add regexp item of whitelist")
	}
	if err := m.addRegexp(m.bcc); err != nil {
		return m, errors.Wrap(err, "in NewMonitor when add regexp item of blaclist")
	}

	if err := m.initialWatcher(); err != nil {
		return m, err
	}
	m.messgage = make(chan message, 100)
	return m, nil
}

func (m *Monitor) addRegexp(cc cacheCount) error {
	if m.regexp == nil {
		m.regexp = make(map[string]*regexp.Regexp)
	}
	for t := range cc {
		if _, ok := m.regexp[t.Regexp]; !ok {
			r, err := regexp.Compile(t.Regexp)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("%q can't compile", t.Regexp))
			}
			m.regexp[target.Regexp] = r
		}
	}
	return nil
}

func (m *Monitor) initialWatcher() error {
	var err error
	m.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "in initialWatcher")
	}
	if err = m.initialProc(); err != nil {
		return errors.Wrap(err, "in initialWatcher")
	}
	return nil
}

// func (monitor *Monitor) matchPattern(str, pattern string) (bool, error) {
// 	re, ok := monitor.regexp[pattern]
// 	if !ok {
// 		return false, fmt.Errorf("not found: regexp of %s", pattern)
// 	}
// 	return matchPattern(str, re), nil
// }

// func matchPattern(str string, r *regexp.Regexp) bool {
// 	return r.Match([]byte(str))
// }
