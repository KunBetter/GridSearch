// rect
package GridSearch

import (
	"strconv"
)

type rect struct {
	Left, Top, Right, Bottom int32
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
	p.lo = (r.Left + r.Right) / 2
	p.la = (r.Top + r.Bottom) / 2
	return p
}

func (r *rect) legal() (ok bool) {
	ok = true
	if r.Left > r.Right {
		ok = false
	}
	if r.Top < r.Bottom {
		ok = false
	}
	return
}

func (r *rect) width() int32 {
	return r.Right - r.Left
}

func (r *rect) height() int32 {
	return r.Top - r.Bottom
}

func (r *rect) equal(other *rect) bool {
	if r.Left == other.Left &&
		r.Top == other.Top &&
		r.Right == other.Right &&
		r.Bottom == other.Bottom {
		return true
	}
	return false
}

func (r *rect) intersection(other *rect) (nr rect, ok bool) {
	nr.Left = max(r.Left, other.Left)
	nr.Top = min(r.Top, other.Top)
	nr.Right = min(r.Right, other.Right)
	nr.Bottom = max(r.Bottom, other.Bottom)
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
		nr.Left = r.Left
		nr.Top = r.Top
		nr.Right = center.lo
		nr.Bottom = center.la
	}

	if qd == 1 {
		nr.Left = center.lo
		nr.Top = r.Top
		nr.Right = r.Right
		nr.Bottom = center.la
	}

	if qd == 2 {
		nr.Left = r.Left
		nr.Top = center.la
		nr.Right = center.lo
		nr.Bottom = r.Bottom
	}

	if qd == 3 {
		nr.Left = center.lo
		nr.Top = center.la
		nr.Right = r.Right
		nr.Bottom = r.Bottom
	}
	return
}

func gridColNum() int32 {
	return (CHINA_RECT.Right-CHINA_RECT.Left)/GRID_TOP_WIDTH + 1
}

func gridRowNum() int32 {
	return (CHINA_RECT.Top-CHINA_RECT.Bottom)/GRID_TOP_HEIGHT + 1
}

func getGridRowCol(lo, la int32) (row, col int32, ok bool) {
	ok = true
	row = (CHINA_RECT.Top - la) / GRID_TOP_HEIGHT
	if row < 0 || row >= GRID_ROW_NUM {
		ok = false
	}
	col = (lo - CHINA_RECT.Left) / GRID_TOP_WIDTH
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
