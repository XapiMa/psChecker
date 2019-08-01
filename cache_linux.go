package psmonitor

type cache map[int]cacheItem
type cacheItem struct {
	t []*Target
	p proc
}

type cacheCount map[*Target]int
