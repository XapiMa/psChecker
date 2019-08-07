package psmonitor

import (
	"fmt"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// Monitor is struct of monitor command
type Monitor struct {
	wcc        cacheCount
	bcc        cacheCount
	cache      cache
	outputPath string
	regexp     map[string]*regexp.Regexp
	watcher    *fsnotify.Watcher
	message    chan message
}
type message struct {
	m    MessageType
	p    Proc
	t    Target
	time time.Time
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

// NewMonitor create new monitor object
func NewMonitor(whitelistPath, blacklistPath, outputPath string) (*Monitor, error) {
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

	m.message = make(chan message, 100)

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
			m.regexp[t.Regexp] = r
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

// Monitor is processes monitoring function
func (m *Monitor) Monitor() error {
	return m.monitor()
}

func (m *Monitor) monitor() error {
	return m.check()
}
