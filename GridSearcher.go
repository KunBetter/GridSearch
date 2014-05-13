// SearchInGrids
package main

import (
	"time"
)

type gridSearher struct {
	gi *gridIndexer
}

func NewGridSearher(gIxer *gridIndexer) *gridSearher {
	return &gridSearher{
		gi: gIxer,
	}
}

func (gs *gridSearher) search(searchRect *rect) []int32 {
	if searchRect.width() > GRID_TOP_WIDTH/4 || searchRect.height() > GRID_TOP_HEIGHT/4 {
		//Search area is too large
		return []int32{}
	}
	resIDsChan := make(chan []int32, 2)
	resIDs := []int32{}
	count := 0
	//top grids for searchRect
	startID := getGridTopIndexKey(searchRect.left, searchRect.top)
	endID := getGridTopIndexKey(searchRect.right, searchRect.bottom)
	if startID == endID {
		count++
		go gs.searchInTopGrid(resIDsChan, startID, searchRect)
	} else {
		sr := startID / GRID_COL_NUM
		sc := startID % GRID_COL_NUM

		er := endID / GRID_COL_NUM
		ec := endID % GRID_COL_NUM
		for i := sr; i <= er; i++ {
			for j := sc; j <= ec; j++ {
				gid := i*GRID_COL_NUM + j
				count++
				go gs.searchInTopGrid(resIDsChan, gid, searchRect)
			}
		}
	}
L:
	for {
		select {
		case res := <-resIDsChan:
			resIDs = append(resIDs, res...)
			count--
			if 0 == count {
				break L
			}
		case <-time.After(time.Second * 3):
			break L
		}
	}

	return resIDs
}

//Search in the corresponding top-level grid, sr is the request rect
func (gs *gridSearher) searchInTopGrid(resIDsChan chan<- []int32, topGridID int32, sr *rect) {
	/*
		Search algorithm:
		Retrieval grid sub-grid quarter ever and seek common ground,
		until the sub-grid cell is equal to the target.
		Save all subnets grid id number;
		Finally, access to data contained in this grid list by id.

		To obtain the corresponding spatial data grid array subscript.
		Each top-level grid corresponds to a spatial index,
		independent of each other.
	*/
	resGridIDs := make(map[int][]int32)
	//***************************************
	curGridRect := gs.gi.gridTopArray[topGridID].pRect
	intersectRect, ok := sr.intersection(curGridRect)
	if ok {
		gs.searchInGrid(resGridIDs, curGridRect, &intersectRect, 0, 0)
	}
	if len(resGridIDs) == 0 {
		return
	}
	/*
		Obtained after quadtree mesh id number,
		get it corresponds to the underlying grid array id,
		you can get indexed.

		Underlying grid computing array obtained in two categories:
		1. the id of the array returned directly to the bottom of the grid,
		will also organize a form of id range.
		2. Id number corresponding to the bottom of the grid interval.
	*/
	bottomGridIDs := []int32{}
	for k, v := range resGridIDs {
		if k != TREEDEPTH-1 {
			for _, vid := range v {
				bottomGridIDs = append(bottomGridIDs, getBottomIDs(vid, k)...)
			}
		}
	}
	tids, ok := resGridIDs[TREEDEPTH-1]
	if ok {
		bottomGridIDs = append(bottomGridIDs, split(tids)...)
	}

	mids := []int32{}
	mids = append(mids, gs.gi.memIxr[topGridID].searchInMem(bottomGridIDs)...)
	mids = append(mids, gs.gi.memIxr[topGridID].searchInSegments(bottomGridIDs)...)
	resIDsChan <- mids
}

//The split in the form of an array of continuous interval
func split(v []int32) []int32 {
	qj := []int32{}
	s, e := 0, 0
	for i := 0; i < len(v)-1; i++ {
		offset := v[i+1] - v[i]
		if offset == 1 {
			e = i + 1
		} else {
			qj = append(qj, v[s])
			qj = append(qj, v[e])
			s = i + 1
			e = i + 1
		}
	}
	qj = append(qj, v[s])
	qj = append(qj, v[e])
	return qj
}

/*
	Obtain the corresponding underlying grid array
	by the current grid id number and the number of layers.
	Improvements:
		For non-mesh bottom, the bottom of the grid corresponding to
		the calculation of the array.
		Only need to return to the range, because it corresponds to
		the id of the underlying grid Is continuous.
*/
func getBottomIDs(id int32, layer int) []int32 {
	se := make([]int32, 2)
	se[0] = id
	se[1] = id
	depth := TREEDEPTH - 1 - layer
	for i := 0; i < depth; i++ {
		se[0] = se[0]*4 + 1
		se[1] = se[1]*4 + 4
	}
	return se
}

func (gs *gridSearher) searchInGrid(resGridIDs map[int][]int32, nodeRect, sr *rect, id int32, layer int) {
	if ok := nodeRect.equal(sr); ok {
		/*
			If this rectangle is equal to retrieve a rectangle,
			You do not need to continue searching quarter.
			Subsequent improvements may consider relaxing this condition.
		*/
		resGridIDs[layer] = append(resGridIDs[layer], id)
		return
	}
	if layer >= TREEDEPTH-1 {
		resGridIDs[layer] = append(resGridIDs[layer], id)
		return
	}
	var i int32 = 0
	for i = 0; i < 4; i++ {
		tr := nodeRect.getQDRect(i)
		ir, ok := tr.intersection(sr)
		if ok {
			//Intersects with the child node, continue down search
			gs.searchInGrid(resGridIDs, &tr, &ir, id*4+1+i, layer+1)
		}
	}
}
