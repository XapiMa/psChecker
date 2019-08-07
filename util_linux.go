package psmonitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

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

func isProcDir(path string) bool {
	path = filepath.Clean(path)
	dir, base := filepath.Split(path)

	if filepath.Clean(dir) != "/proc" {
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
	if len(p) != 4 {
		return false
	}
	if p[1] == "proc" && isNum(p[2]) && p[3] == "fd" {
		return true
	}
	return false
}
func isFd(path string) bool {
	path = filepath.Clean(path)
	p := strings.Split(path, "/")
	if len(p) != 5 {
		return false
	}
	if p[1] == "proc" && isNum(p[2]) && p[3] == "fd" && isNum(p[4]) {
		return true
	}
	return false
}

func getPidFromProcPath(path string) (int, error) {
	path = filepath.Clean(path)
	p := strings.Split(path, "/")
	if len(p) < 3 {
		return 0, fmt.Errorf("%s is not proc path", path)
	}
	if p[1] != "proc" {
		return 0, fmt.Errorf("%s is not proc path", path)
	}
	pid, err := strconv.Atoi(p[2])
	if err != nil {
		return 0, fmt.Errorf("%s is not proc path", path)
	}
	return pid, nil
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

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
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

func parseTypes(typesString string) (int, error) {
	typeStrings := strings.Split(typesString, "|")
	types := 0
	for i, typeString := range typeStrings {
		switch typeString {
		case execSentence:
			types |= execFlag
		case cmdSentence:
			types |= cmdFlag
		case openSentence:
			types |= openFlag
		case userSentence:
			types |= userFlag
		case pidSentence:
			types |= pidFlag
		case "":
		default:
			return types, fmt.Errorf("%dth type is invalid", i+1)
		}
	}
	return types, nil
}
