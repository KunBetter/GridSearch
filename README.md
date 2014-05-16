GridSearch
==========
real-time grid search engine

Installation:
-----
```
	go get github.com/KunBetter/GridSearch
	$GOPATH/bin/GridSearch
```	
Now you have GridSearch running on port 8888 and can use the index and search interfaces as below.

index interface:
-----
```
	http://localhost:8888/index?pt=lo,la,id
	curl -XPUT http://localhost:8888/index -d pt=lo,la,id
```
search interface:
-----
```
	http://localhost:8888/search(random rect)
	http://localhost:8888/search?rect=left,top,right,bottom
	http://localhost:8888/search?rect=9333748,3517838,9381410,3482092
	curl -XPUT http://localhost:8888/search -d rect=left,top,right,bottom
	just for china.
```
Algorithms:
-----
		Assuming there are 10 layers of the quadtree,the num of the bottom grids is 4^9.  
	Each data with latitude and longitude is mapped to the bottom grid.  
	Search when hit the grid at bottom,calculate the corresponding underlying grid array,  
	return the appropriate result.
Features:
-----
	1.A real-time search and indexing incremental updates. 
	2.the first level of the index is designed to save two modes:files, memory.  
		second model search faster, but consumes memory.
Segment control:
-----
	1.Recycling spare segments.
	2.If all segments are in use, increase segment num.
	3.current segments were Recycled stored in a minimum heap.
	4.Adding TTL while improving.for the segment is not used for a long time,  
		some time to delete.
Features:
-----
	1.A real-time search space data.

Performance Monitoring:
-----
	1.go metrics.
Scale:
-----
	1.Spatial data is about 270 million data to about 1G size file,if just store the ID(int32).
next to do:
-----
	1.Infrastructure Optimization
	2.the Consistency problem between index merge operation and the real-time search, because   
	the index merge action will delete the original index files, generate a new index file.  
	response to real time search request,the engine may open index file that have been removed.
