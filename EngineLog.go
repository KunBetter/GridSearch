// EngineLog
package main

import (
	"strconv"
	"strings"
	"time"
)

/*
	Save each data before index for Recovery.
*/

const LOGMAXSIZT = 1024 * 1024 * 4
const LOGFILE = "engine.log"
const LOGFILEINDEX = "engine.log.index"
const LOGSEG = "engine.log.seg"

type EngineLog struct {
	LogFName string
	Length   int
	LogChan  chan *gridData
}

func NewEngineLog() *EngineLog {
	el := &EngineLog{
		LogFName: "engine.1.log",
		Length:   0,
		LogChan:  make(chan *gridData, 100),
	}
	if isFileExist(LOGSEG) {
		buf, _ := readFile2Bytes(LOGSEG)
		el.LogFName = string(buf[:])
		el.Length = int(getFileLength(el.LogFName)) / 12
	} else {
		writeBufToFile(LOGSEG, []byte(el.LogFName))
	}
	go el.process()
	return el
}

func ChangeLogName(fn string) string {
	strs := strings.Split(fn, ".")
	num, err := strconv.Atoi(strs[1])
	if err != nil {
		return fn + ".ex"
	}
	return strs[0] + strconv.Itoa(num+1) + strs[2]
}

func (el *EngineLog) CompressLog(f2compress string) {
	buf, ok := readFile2Bytes(f2compress)
	if !ok {
		print("read " + f2compress + "error.")
		return
	}
	dst, ok := Compress(buf)
	if !ok {
		print("compress " + f2compress + "error.")
		return
	}
	logfLen := int32(getFileLength(LOGFILE))
	writeBufAppendFile(LOGFILE, dst)
	writeBufAppendFile(LOGFILEINDEX, Int32ToBytes(logfLen))
	rmFile(f2compress)
}

func (el *EngineLog) process() {
	for {
		select {
		case gd := <-el.LogChan:
			buf := []byte{}
			buf = append(buf, Int32ToBytes(gd.lo)...)
			buf = append(buf, Int32ToBytes(gd.la)...)
			buf = append(buf, Int32ToBytes(gd.id)...)
			el.Length++
			writeBufAppendFile(el.LogFName, buf)
			if el.Length >= LOGMAXSIZT {
				//first,change LogFName,to store the subsequent data
				f2compress := el.LogFName
				el.LogFName = ChangeLogName(el.LogFName)
				el.Length = 0
				writeBufToFile(LOGSEG, []byte(el.LogFName))
				//compress log
				go el.CompressLog(f2compress)
			}
		case <-time.Tick(time.Hour):
			//add each hour flag
			buf := []byte{}
			buf = append(buf, Int32ToBytes(0)...)
			buf = append(buf, Int32ToBytes(0)...)
			buf = append(buf, Int32ToBytes(0)...)
			el.Length++
			writeBufAppendFile(el.LogFName, buf)
		}
	}
}

func (el *EngineLog) LogData(gd *gridData) {
	el.LogChan <- gd
}
