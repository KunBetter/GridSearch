// GridIndices
package main

import (
	"github.com/rcrowley/go-metrics"
	"log"
	"os"
)

type gridIndexer struct {
	gridTopArray  []gridTop       //top grid array
	InputDataFlow chan []gridData //Data Index Flow
	memIxr        []*memIndexer   //Memory Index
	indexMeter    metrics.Meter   //Indexing speed monitoring
	el            *EngineLog      //Engine backup system
}

//func make([]T, len, cap) []T
func NewGridIndexer() *gridIndexer {
	gi := gridIndexer{
		gridTopArray:  make([]gridTop, gridNum, gridNum),
		InputDataFlow: make(chan []gridData, 1000),
		memIxr:        make([]*memIndexer, gridNum, gridNum),
		indexMeter:    metrics.NewMeter(),
		el:            NewEngineLog(),
	}
	//Divide the top grid
	var i int32 = 0
	for i = 0; i < gridNum; i++ {
		tRow := i / GRID_COL_NUM
		tCol := i % GRID_COL_NUM
		gi.gridTopArray[i].pRect = &rect{
			chinaRect().Left + GRID_TOP_WIDTH*tCol,
			chinaRect().Top - GRID_TOP_HEIGHT*tRow,
			chinaRect().Left + GRID_TOP_WIDTH*tCol + GRID_TOP_WIDTH,
			chinaRect().Top - GRID_TOP_HEIGHT*tRow - GRID_TOP_HEIGHT,
		}
		//Memory indexed array
		gi.memIxr[i] = NewMemIndexer(i)
	}
	metrics.Register("indexing", gi.indexMeter)
	return &gi
}

func (gi *gridIndexer) close() {
	close(gi.InputDataFlow)
	var i int32 = 0
	for i = 0; i < gridNum; i++ {
		gi.memIxr[i].close()
	}
}

func (gi *gridIndexer) worker(inFlow <-chan []gridData) {
	for data := range inFlow {
		for _, ix := range data {
			//Indexers single data
			gi.el.LogData(&ix)
			gridTopID := getGridTopIndexKey(ix.lo, ix.la)
			if gridTopID < 0 || gridTopID >= gridNum {
				continue
			}
			if gridTopID != -1 {
				bid := gi.getBottomGridID(gridTopID, &ix)
				gi.indexMeter.Mark(1)
				gi.memIxr[gridTopID].dataFlow <- mData{ix.id, bid}
			}
		}
	}
}

func (gi *gridIndexer) indexing() {
	go metrics.Log(metrics.DefaultRegistry, 60e9, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	for i := 0; i < INDEXTHREADNUM; i++ {
		go gi.worker(gi.InputDataFlow)
	}
}

func (gi *gridIndexer) indexDocs(pts []gridData) {
	gi.InputDataFlow <- pts
}

/*
	The point is mapped to the underlying quadtree mesh
	and returns the underlying grid quadtree id number
*/
func (gi *gridIndexer) getBottomGridID(topGridID int32, d *gridData) int32 {
	var layer, curID int32 = 0, 0
	return get4TreeBottomGridID(gi.gridTopArray[topGridID].pRect,
		&point{d.lo, d.la}, layer, curID)
}

func get4TreeBottomGridID(curRect *rect, p *point, curLayer, curID int32) int32 {
	if curLayer >= TREEDEPTH-1 {
		return curID
	}
	qd := curRect.getQD(p)
	nextRect := curRect.getQDRect(qd)
	return get4TreeBottomGridID(&nextRect, p, curLayer+1, curID*4+1+qd)
}
