package stats

import (
	"time"
)

var es *EngineStats

func StatsInit() {
	es = NewEngineStats()
}

func SearchIn() {
	es.search.ctv <- NewTimedValue(time.Now(), 1)
}

func IndexIn() {
	es.index.ctv <- NewTimedValue(time.Now(), 1)
}

func Stats() *DurationCounter {
	return es.search.dc
}

type statistics struct {
	dc  *DurationCounter
	ctv chan *TimedValue
}

func NewStatistics() *statistics {
	return &statistics{
		dc:  NewDurationCounter(),
		ctv: make(chan *TimedValue, 100),
	}
}

type EngineStats struct {
	search *statistics
	index  *statistics
}

func NewEngineStats() *EngineStats {
	es := &EngineStats{
		search: NewStatistics(),
		index:  NewStatistics(),
	}
	go es.process()
	return es
}

func (es *EngineStats) process() {
	for {
		select {
		case tv := <-es.search.ctv:
			es.search.dc.Add(tv)
		case tv := <-es.index.ctv:
			es.index.dc.Add(tv)
		}
	}
}
