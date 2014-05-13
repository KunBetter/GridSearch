// rect
package main

import (
	"strconv"
)

type rect struct {
	left, top, right, bottom int32
}

func NewRectBy4String(v []string) (*rect, bool) {
	tleft, err := strconv.Atoi(v[0])
	if err != nil {
		return nil, false
	}
	ttop, err := strconv.Atoi(v[1])
	if err != nil {
		return nil, false
	}
	tright, err := strconv.Atoi(v[2])
	if err != nil {
		return nil, false
	}
	tbottom, err := strconv.Atoi(v[3])
	if err != nil {
		return nil, false
	}
	r := &rect{
		int32(tleft),
		int32(ttop),
		int32(tright),
		int32(tbottom),
	}
	return r, r.legal()
}

func (r *rect) getCenter() point {
	var p point
	p.lo = (r.left + r.right) / 2
	p.la = (r.top + r.bottom) / 2
	return p
}

func (r *rect) legal() (ok bool) {
	ok = true
	if r.left > r.right {
		ok = false
	}
	if r.top < r.bottom {
		ok = false
	}
	return
}

func (r *rect) width() int32 {
	return r.right - r.left
}

func (r *rect) height() int32 {
	return r.top - r.bottom
}

func (r *rect) equal(other *rect) bool {
	if r.left == other.left &&
		r.top == other.top &&
		r.right == other.right &&
		r.bottom == other.bottom {
		return true
	}
	return false
}

func (r *rect) intersection(other *rect) (nr rect, ok bool) {
	nr.left = max(r.left, other.left)
	nr.top = min(r.top, other.top)
	nr.right = min(r.right, other.right)
	nr.bottom = max(r.bottom, other.bottom)
	ok = nr.legal()
	return
}

func (r *rect) contain(other *rect) (nr rect, ok bool) {
	nr, ok = r.intersection(other)
	if ok {
		ok = nr.equal(other)
	}
	return
}

func (r *rect) getQD(p *point) int32 {
	center := r.getCenter()
	var qd int32 = 0
	if p.lo > center.lo {
		qd += 1
	}
	if p.la < center.la {
		qd += 2
	}
	return qd
}

func (r *rect) getQDRect(qd int32) (nr rect) {
	center := r.getCenter()

	if qd == 0 {
		nr.left = r.left
		nr.top = r.top
		nr.right = center.lo
		nr.bottom = center.la
	}

	if qd == 1 {
		nr.left = center.lo
		nr.top = r.top
		nr.right = r.right
		nr.bottom = center.la
	}

	if qd == 2 {
		nr.left = r.left
		nr.top = center.la
		nr.right = center.lo
		nr.bottom = r.bottom
	}

	if qd == 3 {
		nr.left = center.lo
		nr.top = center.la
		nr.right = r.right
		nr.bottom = r.bottom
	}
	return
}

func gridColNum() int32 {
	return (CHINA_RECT.right-CHINA_RECT.left)/GRID_TOP_WIDTH + 1
}

func gridRowNum() int32 {
	return (CHINA_RECT.top-CHINA_RECT.bottom)/GRID_TOP_HEIGHT + 1
}

func getGridRowCol(lo, la int32) (row, col int32, ok bool) {
	ok = true
	row = (CHINA_RECT.top - la) / GRID_TOP_HEIGHT
	if row < 0 || row >= GRID_ROW_NUM {
		ok = false
	}
	col = (lo - CHINA_RECT.left) / GRID_TOP_WIDTH
	if col < 0 || col >= GRID_COL_NUM {
		ok = false
	}

	return
}

func getGridTopIndexKey(lo, la int32) int32 {
	row, col, ok := getGridRowCol(lo, la)
	if !ok {
		return -1
	}

	return row*GRID_COL_NUM + col
}
