// Utils
package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"sort"
	"syscall"
)

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

//calculate the square belongs to which layer of the quadtree by id.
func getLayerByID(id int32) float64 {
	var f64ID float64 = float64(id)
	layer := math.Floor(math.Log2(f64ID*3+1) / 2)
	return layer
}

//calculate the start index of each layer in quadtree.
func getLayerFirstIx(layer float64) int32 {
	return int32((math.Pow(4, layer) - 1) / 3)
}

func getLayerLastIx(layer float64) int32 {
	return int32((math.Pow(4, layer) - 1) * 4 / 3)
}

/*
	formula to calculate the final grid index for Each layer:
	f(n)=(4^n-1)*4/3
*/
func isLayerLastIx(id int32) bool {
	var f64ID float64 = float64(id)
	layer := math.Floor(math.Log2(f64ID*3+1) / 2)
	lastID := int32((math.Pow(4, layer+1) - 4) / 3)
	if lastID == id {
		return true
	}
	return false
}

func path(topGridID int32) string {
	return fmt.Sprintf("%s/grid%d.", indexDir, topGridID)
}

func sortedKeys(m map[int32]*SegmentAttr) []int32 {
	var sm sortedMap
	sm.m = m
	sm.k = make([]int32, len(m))
	i := 0
	for key, _ := range m {
		sm.k[i] = key
		i++
	}
	sort.Sort(&sm)
	return sm.k
}

func Int32ToBytes(i int32) []byte {
	buf := []byte{0, 0, 0, 0}
	binary.LittleEndian.PutUint32(buf, uint32(i))
	return buf
}

func BytesToInt32(buf []byte) int32 {
	return int32(binary.LittleEndian.Uint32(buf))
}

func i32ToB(v []int32) []byte {
	bLen := len(v) * 4
	buf := make([]byte, bLen)
	offset := 0
	for i := 0; i < len(v); i++ {
		copy(buf[offset:], Int32ToBytes(v[i]))
		offset += 4
	}
	return buf
}

func b2i32(b []byte) []int32 {
	blen := len(b) / 4
	buf := make([]int32, blen)
	tb := []byte{0, 0, 0, 0}
	for i := 0; i < blen; i++ {
		tb[0] = b[i*4]
		tb[1] = b[i*4+1]
		tb[2] = b[i*4+2]
		tb[3] = b[i*4+3]
		buf[i] = BytesToInt32(tb)
	}
	return buf
}

//*******************FILE API*********************

func check(e error) {
	if e != nil {
		panic(e)
		print(e)
	}
}

func createDir(dir string) {
	if !isDirExist(dir) {
		err := os.Mkdir(dir, 0755)
		check(err)
	}
}

func isDirExist(dir string) bool {
	f, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	} else {
		return f.IsDir()
	}
}

func readFile(fn string) ([]int32, bool) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, false
	} else {
		return b2i32(b), true
	}
}

func readFile2Bytes(fn string) ([]byte, bool) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, false
	} else {
		return b, true
	}
}

func isFileExist(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil || os.IsExist(err)
}

func rmFile(fn string) {
	if !isFileExist(fn) {
		return
	}
	err := os.Remove(fn)
	if err != nil {
		panic("rm file error." + fn)
	}
}

func writeBufToFile(fn string, buffer []byte) {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)
	defer f.Close()
	bLen, err := f.Write(buffer)
	check(err)
	if bLen <= 0 {
		print("write file error!", fn)
	}
}

func getFileLength(fn string) int64 {
	fi, err := os.Stat(fn)
	if err != nil {
		panic("can not get the file length!")
	}
	return fi.Size()
}

//*********INTERRUPT***************************
func OnInterrupt(fn func()) {
	// deal with control+c,etc
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for _ = range signalChan {
			fn()
			os.Exit(0)
		}
	}()
}
