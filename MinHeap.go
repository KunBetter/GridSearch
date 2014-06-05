// minheap
package GridSearch

/*
1.the first node in the array is spare.
2.in accordance with the array subscript,
the subscript of n nodes, its children nodes subscript are 2*n and 2*n+1;
3.inserting,the node is inserted into the end of the array,
then adjust the heap.
4.delete the smallest root node,
the root and last nodes exchanged, and then adjust the heap.
*/

type MinHeap struct {
	mhLen, length int
	mhA           []int32
}

func NewMinHeap(size int) *MinHeap {
	return &MinHeap{
		mhLen:  0,
		length: size,
		mhA:    make([]int32, size+1),
	}
}

func (mh *MinHeap) push(e int32) {
	mh.lazyGrow()
	if mh.mhLen == mh.length {
		if e <= mh.mhA[1] {
			return
		}
		mh.pop()
	}
	mh.mhLen++
	if mh.mhLen == 1 {
		mh.mhA[mh.mhLen] = e
		return
	}
	i := mh.mhLen
	for mh.mhA[i/2] > e {
		mh.mhA[i] = mh.mhA[i/2]
		i = i / 2
	}
	mh.mhA[i] = e
}

func (mh *MinHeap) pop() (e int32) {
	if mh.mhLen <= 0 {
		e = -1
		return
	}
	e = mh.mhA[1]
	lastE := mh.mhA[mh.mhLen]
	mh.mhA[mh.mhLen] = 0
	mh.mhLen--
	if mh.mhLen == 0 {
		return
	}
	i := 1
	for i*2 <= mh.mhLen {
		i = i * 2
		/*
			compared with the min subnode.
			子节点必须比父节点大，不然只比较一个，
			就会出现父节点比一个子节点小，
			但是比另一个大的情况
			Child nodes must be greater than parent node,
			or just comparing one,parent node will lesser
			than a small child but larger than the other.
		*/
		if i+1 <= mh.mhLen && mh.mhA[i] > mh.mhA[i+1] {
			i++
		}
		if lastE > mh.mhA[i] {
			mh.mhA[i/2] = mh.mhA[i]
		} else {
			i = i / 2
			break
		}
	}
	mh.mhA[i] = lastE
	return
}

func (mh *MinHeap) lazyGrow() {
	if mh.full() {
		mh.grow()
	}
}

func (mh *MinHeap) len() int {
	return mh.mhLen
}

func (mh *MinHeap) full() bool {
	return mh.mhLen == mh.length
}

func (mh *MinHeap) grow() {
	big := make([]int32, mh.length*2+1)
	for i := 1; i <= mh.mhLen; i++ {
		big[i] = mh.mhA[i]
	}
	mh.mhA = big
	mh.length *= 2
}
