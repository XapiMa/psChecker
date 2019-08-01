package psmonitor

import (
	"io/ioutil"
	"sort"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func readList(path string) (map[*Target]int, error) {
	m := make(map[*Target]int)

	list, err := parseConfigYml(path)
	if err != nil {
		return m, errors.Wrapf(err, "in readList: for %s", path)
	}
	for i := range list {
		sortStringSlice(list[i].Open)
	}
	for i := range list {
		m[&list[i]] = 0
	}
	return m, nil
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

func sortStringSlice(s []string) {
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
}
