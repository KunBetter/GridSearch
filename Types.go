// Types
package main

type gridData struct {
	lo, la int32
	id     int32 //Unique number
}

//top grids
type gridTop struct {
	pRect *rect //the rect to each top grid
}

type point struct {
	lo, la int32
}

func chinaRect() rect {
	return rect{
		left:   7181646,
		top:    5616041,
		right:  13641237,
		bottom: 195187,
	}
}

type mData struct {
	id, bid int32
}

//segment attribute
type SegmentAttr struct {
	size    int32
	merging bool
}

type sortedMap struct {
	m map[int32]*SegmentAttr
	k []int32
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.k[i]].size < sm.m[sm.k[j]].size
}

func (sm *sortedMap) Swap(i, j int) {
	sm.k[i], sm.k[j] = sm.k[j], sm.k[i]
}