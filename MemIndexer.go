// MemIndexer
package GridSearch

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type memIndexer struct {
	memIXData   []map[int32][]int32 //Memory index data structure for real-time search
	curMem      int                 //0,1 The current map for memory index
	counter     int                 //Memory data counter
	topGridID   int32               //top grid id
	dataFlow    chan mData          //data flow
	done        chan int            //Write files have completed
	swap        bool                //can swap the mem map
	nextSeg     *MinHeap            //Available segment
	mergingChan chan bool           //Combined index signal channel
	/*
		The maximum number of index files,
		as more and more documents,
		this value also increases
	*/
	segmentUpper int32
	firstIXs     map[int32][]int32 //first indices
	segment      struct {
		sync.RWMutex
		attr map[int32]*SegmentAttr //Segment attributes
	}
}

func NewMemIndexer(tid int32) *memIndexer {
	mi := memIndexer{
		memIXData:    make([]map[int32][]int32, 2, 2),
		firstIXs:     make(map[int32][]int32),
		counter:      0,
		curMem:       0,
		topGridID:    tid,
		dataFlow:     make(chan mData),
		done:         make(chan int),
		swap:         true,
		segmentUpper: 0,
		nextSeg:      NewMinHeap(10),
		mergingChan:  make(chan bool),
	}
	mi.segment.attr = make(map[int32]*SegmentAttr)
	mi.memIXData[0] = make(map[int32][]int32)
	mi.memIXData[1] = make(map[int32][]int32)

	if FIRSTINDEXMODE == FIRSTINDEXMEM {
		mi.loadFirstIXs()
	}
	mi.loadSnapshot()

	go mi.addData()
	go mi.process()

	return &mi
}

func (mi *memIndexer) loadFirstIXs() bool {
	tPath := path(mi.topGridID)
	fPath := fmt.Sprintf("%sfir", tPath)
	exist := isFileExist(fPath)
	if !exist {
		print(fPath, "does not exist.")
		return true
	}
	fIXsInt32s, ok := readFile(fPath)
	if !ok {
		return false
	}
	t := int32(len(fIXsInt32s)) / (BOTTOMGRIDNUM + 1)
	var offset, i int32 = 0, 0
	for i = 0; i < t; i++ {
		offset = i * (BOTTOMGRIDNUM + 1)
		mi.firstIXs[fIXsInt32s[offset]] = fIXsInt32s[1+offset : BOTTOMGRIDNUM+1+offset]
	}
	return true
}

/*
	Persistence save first index
	store the first index:
		when searching, the first index readed in memory,
		so first index File mirroring only need to synchronize
		with the second indexes.
*/
func (mi *memIndexer) storeFirstIXs() bool {
	tPath := path(mi.topGridID)
	buf := []byte{}
	for k, v := range mi.firstIXs {
		buf = append(buf, Int32ToBytes(k)...)
		buf = append(buf, i32ToB(v)...)
	}
	fPath := fmt.Sprintf("%sfir", tPath)
	rmFile(fPath)
	writeBufToFile(fPath, buf)
	return true
}

func (mi *memIndexer) loadSnapshot() bool {
	tPath := path(mi.topGridID)
	path := fmt.Sprintf("%ssnapshot", tPath)
	exist := isFileExist(path)
	if !exist {
		print(path, "does not exist.")
		return true
	}
	segAttr, ok := readFile(path)
	if !ok {
		return false
	}
	alen := segAttr[0]
	var i int32 = 0
	for i = 0; i < alen; i++ {
		mi.segment.attr[segAttr[i*2+1]] = &SegmentAttr{
			size:    segAttr[i*2+2],
			merging: false,
		}
	}
	mhLen := segAttr[alen+1]
	for i = 0; i < mhLen; i++ {
		mi.nextSeg.push(segAttr[alen+2+i])
	}
	mi.segmentUpper = segAttr[alen+1+mhLen+1]
	return true
}

func (mi *memIndexer) storeSnapshot() {
	tPath := path(mi.topGridID)
	//*********SEGMENT*********************
	buf := []byte{0}
	for k, v := range mi.segment.attr {
		buf = append(buf, Int32ToBytes(k)...)
		buf = append(buf, Int32ToBytes(v.size)...)
	}
	alen := int32(len(buf))/4 - 1
	buf = append(buf, Int32ToBytes(alen)...)
	//*****MINHEAP***********************
	buf = append(buf, Int32ToBytes(int32(mi.nextSeg.mhLen))...)
	for i := 0; i < mi.nextSeg.mhLen; i++ {
		buf = append(buf, Int32ToBytes(mi.nextSeg.mhA[i])...)
	}
	//*****segmentUpper***********************
	buf = append(buf, Int32ToBytes(mi.segmentUpper)...)

	path := fmt.Sprintf("%ssnapshot", tPath)
	rmFile(path)
	writeBufToFile(path, buf)
}

func (mi *memIndexer) close() {
	<-time.After(time.Second * 1)
	if FIRSTINDEXMODE == FIRSTINDEXMEM {
		mi.storeFirstIXs()
	}
	mi.storeMem(mi.nextSegment(true))
	mi.storeSnapshot()
}

func (mi *memIndexer) addData() {
	for {
		select {
		case md := <-mi.dataFlow:
			mi.memIXData[mi.curMem][md.bid] = append(mi.memIXData[mi.curMem][md.bid], md.id)
			mi.counter++
			if mi.counter >= DOCMERGENUM && mi.swap {
				mi.swap = false
				go mi.mem2File(mi.curMem, mi.nextSegment(false))
				mi.curMem = (mi.curMem + 1) % 2
				mi.counter = 0
			}
		case <-mi.done:
			mi.swap = true
		}
	}
}

func (mi *memIndexer) process() {
	for {
		select {
		case <-mi.mergingChan:
			go func() {
				seg1, seg2, ok := mi.merge2Segs()
				if ok {
					mi.ixMerge(seg1, seg2, mi.nextSegment(true))
				}
			}()
		}
	}
}

func (mi *memIndexer) merge2Segs() (seg1, seg2 int32, ok bool) {
	mi.segment.Lock()
	defer mi.segment.Unlock()
	ok = false
	tSegs := sortedKeys(mi.segment.attr)
	segs := make([]int32, 2)
	i := 0
	for _, k := range tSegs {
		if !mi.segment.attr[k].merging {
			segs[i] = k
			i++
			if i >= 2 {
				break
			}
		}
	}
	if i <= 1 {
		seg1 = -1
		seg2 = -1
		return
	}
	seg1 = segs[0]
	seg2 = segs[1]
	if mi.segment.attr[seg1].size >= 134217728 {
		return
	}
	mi.segment.attr[seg1].merging = true
	mi.segment.attr[seg2].merging = true
	ok = true
	runtime.Gosched()
	return
}

/*
	Can not be blocked.
*/
func (mi *memIndexer) nextSegment(merging bool) (ns int32) {
	qlen := int32(mi.nextSeg.len())

	if qlen <= mi.segmentUpper/3 && mi.segmentUpper >= 2 && !merging {
		mi.mergingChan <- true
	}

	if qlen <= 0 {
		mi.segmentUpper++
		ns = mi.segmentUpper
	} else {
		ns = mi.nextSeg.pop()
	}

	return
}

func (mi *memIndexer) searchInMem(bids []int32) []int32 {
	mRes := []int32{}
	blen := len(bids) / 2
	for i := 0; i < blen; i++ {
		var k, s, e int32 = 0, bids[i*2], bids[i*2+1]
		for k = s; k <= e; k++ {
			ids, ok := mi.memIXData[0][k]
			if ok {
				mRes = append(mRes, MergeSort(ids)...)
			}
			ids, ok = mi.memIXData[1][k]
			if ok {
				mRes = append(mRes, MergeSort(ids)...)
			}
		}
	}
	return mRes
}

func (mi *memIndexer) searchInSegments(bids []int32) []int32 {
	mRes := []int32{}
	sa := []int32{}
	mi.segment.Lock()
	defer mi.segment.Unlock()
	for k, _ := range mi.segment.attr {
		sa = append(sa, k)
	}
	runtime.Gosched()
	for _, k := range sa {
		mRes = append(mRes, mi.searchInSeg(k, bids)...)
	}

	return mRes
}

func (mi *memIndexer) searchInSeg(seg int32, bids []int32) []int32 {
	tPath := path(mi.topGridID)
	sPath := fmt.Sprintf("%s%d.sec", tPath, seg)
	fIXs, ok := []int32{}, false
	if FIRSTINDEXMODE == FIRSTINDEXFILE {
		fPath := fmt.Sprintf("%s%d.fir", tPath, seg)
		fIXs, ok = readFile(fPath)
	} else {
		fIXs, ok = mi.firstIXs[seg]
	}
	if !ok {
		return []int32{}
	}
	sIXs, ok := readFile(sPath)
	if !ok {
		return []int32{}
	}
	mRes := []int32{}
	blen := len(bids) / 2
	for i := 0; i < blen; i++ {
		mRes = append(mRes, mi.getResults(fIXs, sIXs, bids[i*2]-BOTTOMFIRSTGRIDID, bids[i*2+1]-BOTTOMFIRSTGRIDID)...)
	}
	return mRes
}

func (mi *memIndexer) getResults(fIXs, sIXs []int32, sbid, ebid int32) []int32 {
	resIDs := []int32{}
	if ebid == BOTTOMGRIDNUM-1 {
		resIDs = append(resIDs, sIXs[fIXs[sbid]:]...)
	} else {
		resIDs = append(resIDs, sIXs[fIXs[sbid]:fIXs[ebid+1]]...)
	}
	return resIDs
}

func (mi *memIndexer) rmIXFile(seg int32) {
	tPath := path(mi.topGridID)
	fPath := fmt.Sprintf("%s%d.fir", tPath, seg)
	sPath := fmt.Sprintf("%s%d.sec", tPath, seg)
	rmFile(fPath)
	rmFile(sPath)
}

func (mi *memIndexer) storeMem(seg int32) {
	bufLen := BOTTOMGRIDNUM
	firstIXs := make([]int32, bufLen, bufLen)
	secondIXs := []int32{}
	var offset int32 = 0
	var i int32 = 0
	for i = 0; i < bufLen; i++ {
		firstIXs[i] = offset
		ids, ok := mi.memIXData[0][i+BOTTOMFIRSTGRIDID]
		if ok {
			ids = MergeSort(ids)
			secondIXs = append(secondIXs, ids...)
			offset += int32(len(ids))
		}
		ids, ok = mi.memIXData[1][i+BOTTOMFIRSTGRIDID]
		if ok {
			ids = MergeSort(ids)
			secondIXs = append(secondIXs, ids...)
			offset += int32(len(ids))
		}
	}
	tPath := path(mi.topGridID)
	sPath := fmt.Sprintf("%s%d.sec", tPath, seg)
	writeBufToFile(sPath, i32ToB(secondIXs))
	fPath := fmt.Sprintf("%s%d.fir", tPath, seg)
	writeBufToFile(fPath, i32ToB(firstIXs))

	mi.segment.Lock()
	defer mi.segment.Unlock()
	mi.segment.attr[seg] = &SegmentAttr{
		offset,
		false,
	}

	runtime.Gosched()
}

/*
	The persistence of memory index to a file.
	empty memory after Persisted storage completed,
	or can not to be a real-time search, will lose data.
*/
func (mi *memIndexer) mem2File(curMem int, seg int32) {
	bufLen := BOTTOMGRIDNUM
	firstIXs := make([]int32, bufLen, bufLen)
	secondIXs := []int32{}
	var offset int32 = 0
	var i int32 = 0
	for i = 0; i < bufLen; i++ {
		firstIXs[i] = offset
		ids, ok := mi.memIXData[curMem][i+BOTTOMFIRSTGRIDID]
		if ok {
			ids = MergeSort(ids)
			secondIXs = append(secondIXs, ids...)
			offset += int32(len(ids))
		}
		mi.memIXData[curMem][i+BOTTOMFIRSTGRIDID] = []int32{}
	}
	tPath := path(mi.topGridID)
	sPath := fmt.Sprintf("%s%d.sec", tPath, seg)
	writeBufToFile(sPath, i32ToB(secondIXs))
	if FIRSTINDEXMODE == FIRSTINDEXFILE {
		fPath := fmt.Sprintf("%s%d.fir", tPath, seg)
		writeBufToFile(fPath, i32ToB(firstIXs))
	} else {
		mi.firstIXs[seg] = firstIXs
		go mi.storeFirstIXs()
	}

	mi.segment.Lock()
	defer mi.segment.Unlock()
	mi.segment.attr[seg] = &SegmentAttr{
		offset,
		false,
	}

	runtime.Gosched()
	mi.done <- curMem
}

/*
	index merge.
*/
func (mi *memIndexer) ixMerge(seg1, seg2 int32, seg int32) {
	tPath := path(mi.topGridID)

	sPath1 := fmt.Sprintf("%s%d.sec", tPath, seg1)
	sIXs1, _ := readFile(sPath1)

	sPath2 := fmt.Sprintf("%s%d.sec", tPath, seg2)
	sIXs2, _ := readFile(sPath2)

	fIXs1, fIXs2 := []int32{}, []int32{}

	if FIRSTINDEXMODE == FIRSTINDEXFILE {
		fPath1 := fmt.Sprintf("%s%d.fir", tPath, seg1)
		fIXs1, _ = readFile(fPath1)

		fPath2 := fmt.Sprintf("%s%d.fir", tPath, seg2)
		fIXs2, _ = readFile(fPath2)
	} else {
		fIXs1 = mi.firstIXs[seg1]
		fIXs2 = mi.firstIXs[seg2]
	}

	bufLen := BOTTOMGRIDNUM
	firstIXs := make([]int32, bufLen, bufLen)
	secondIXs := []int32{}
	var i, mergeLen int32 = 0, 0
	for i = 0; i < bufLen-1; i++ {
		firstIXs[i] = mergeLen
		mergeSIX := Merge(sIXs1[fIXs1[i]:fIXs1[i+1]], sIXs2[fIXs2[i]:fIXs2[i+1]])
		secondIXs = append(secondIXs, mergeSIX...)
		mergeLen += int32(len(mergeSIX))
	}
	firstIXs[i] = mergeLen
	mergeSIX := Merge(sIXs1[fIXs1[i]:], sIXs2[fIXs2[i]:])
	secondIXs = append(secondIXs, mergeSIX...)
	mergeLen += int32(len(mergeSIX))

	sPath := fmt.Sprintf("%s%d.sec", tPath, seg)
	writeBufToFile(sPath, i32ToB(secondIXs))
	if FIRSTINDEXMODE == FIRSTINDEXFILE {
		fPath := fmt.Sprintf("%s%d.fir", tPath, seg)
		writeBufToFile(fPath, i32ToB(firstIXs))
	} else {
		mi.firstIXs[seg] = firstIXs
	}

	//*******************************************
	mi.segment.Lock()
	defer mi.segment.Unlock()
	mi.segment.attr[seg] = &SegmentAttr{
		mi.segment.attr[seg1].size + mi.segment.attr[seg2].size,
		false,
	}
	delete(mi.segment.attr, seg1)
	delete(mi.segment.attr, seg2)

	if FIRSTINDEXMODE == FIRSTINDEXMEM {
		delete(mi.firstIXs, seg1)
		delete(mi.firstIXs, seg2)
	}
	//*******************************************

	mi.rmIXFile(seg1)
	mi.rmIXFile(seg2)

	mi.nextSeg.push(seg1) //Recovery segments
	mi.nextSeg.push(seg2)

	runtime.Gosched()
}
