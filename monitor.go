package pschecker

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Target struct {
	Exec string   `yaml:"exec"`
	Cmd  string   `yaml:"cmd"`
	Open []string `yaml:"open"`
	User string   `yaml:"user"`
	Pid  int      `yaml:"pid"`
}

func (checker *Checker) Monitor(configPath string) error {
	conf, err := parseConfigYml(configPath)
	if err != nil {
		return errors.Wrap(err, "cause in Monitor")
	}
	fmt.Println(conf)

	return nil
}

func parseConfigYml(configPath string) ([]Target, error) {
	targets := make([]Target, 10000)
	errorWrap := func(err error) error {
		return errors.Wrap(err, "cause in parseConfigYml")
	}

	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return targets, errorWrap(err)
	}
	if err := yaml.Unmarshal(buf, &targets); err != nil {
		return targets, errorWrap(err)
	}

	return targets, nil

}
