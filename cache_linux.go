package psmonitor

import "fmt"

type cache map[int]cacheItem
type cacheItem struct {
	t []*Target
	p Proc
}

type cacheCount map[*Target]int

func (c cache) in(pid int) bool {
	_, ok := c[pid]
	return ok

}

func (c cache) add(item cacheItem) error {
	pid := item.p.pid
	if _, ok := c[pid]; ok {
		return fmt.Errorf("pid:%d is already cached as %v. item is %v", pid, c[pid], item)
	}
	c[pid] = item
	return nil
}

func (c cache) del(pid int) error {
	if _, ok := c[pid]; !ok {
		return fmt.Errorf("pid:%d is not cached", pid)
	}
	delete(c, pid)
	return nil
}

func (c cacheCount) in(t *Target) bool {
	count, ok := c[t]
	if ok {
		if count != 0 {
			return true
		}
	}
	return false
}

func (c cacheCount) is(t *Target) bool {
	_, ok := c[t]
	return ok
}

func (c cacheCount) add(t *Target) error {
	if _, ok := c[t]; !ok {
		return fmt.Errorf("%v is not key of cacheCount", *t)
	}
	// もしも 0 → 1 ならメッセージを出す
	c[t]++
	return nil
}
func (c cacheCount) del(t *Target) {
	c[t]--
	// もしも0になったらメッセージを出す
}
