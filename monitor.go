package pschecker

import (
	"github.com/pkg/errors"
)

type Monitor struct {
	outputPath string
	whitelist  []Target
	blacklist  []Target
}

func NewMonitor(whitelistPath, blacklistPath, outputPath string) (*Monitor, error) {
	m := new(Monitor)
	var err error
	m.whitelist, err = parseConfigYml(whitelistPath)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor: for whitelist")
	}
	m.blacklist, err = parseConfigYml(blacklistPath)
	if err != nil {
		return m, errors.Wrap(err, "cause in NewMonitor: for blacklist")
	}

	m.outputPath = outputPath
	return m, nil
}

func (monitor *Monitor) Monitor() error {

	return nil
}
