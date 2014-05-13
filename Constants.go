// Constants
package main

import (
	"fmt"
	"math"
)

/*
	the relationship between parent node and
	child node:
	c1 = p*4 + 1
	c2 = p*4 + 2
	c3 = p*4 + 3
	c4 = p*4 + 4
	0|1,2,3,4|5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20|21,...
	suppose we have 11 layer:
	4^0 + 4^1 + ... + 4^10 = 1398101
*/

const (
	TREEDEPTH       = 11
	GRID_TOP_WIDTH  = 1024 * 512
	GRID_TOP_HEIGHT = 1024 * (256 + 128)
	indexDir        = "data"
	FIRSTINDEXFILE  = 0 //first indices stored in files
	FIRSTINDEXMEM   = 1 //first indices stored in mem
	FIRSTINDEXMODE  = FIRSTINDEXFILE
	INDEXTHREADNUM  = 1     //gorountine num for indexing
	DOCMERGENUM     = 20000 //the threshold of the doc to store in file
)

var (
	print              = fmt.Println
	GRID_COL_NUM       = gridColNum()
	GRID_ROW_NUM       = gridRowNum()
	CHINA_RECT         = chinaRect()
	gridNum            = GRID_ROW_NUM * GRID_COL_NUM
	BOTTOMLEVELGRIDNUM = math.Pow(4, TREEDEPTH-1)
	BOTTOMGRIDNUM      = int32(BOTTOMLEVELGRIDNUM)
	BOTTOMFIRSTGRIDID  = getLayerFirstIx(float64(TREEDEPTH - 1))
)
