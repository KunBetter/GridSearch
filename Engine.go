// Engine
package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Engine struct {
	gi *gridIndexer
	gs *gridSearher
}

func (engine *Engine) start() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	createDir(indexDir)

	engine.gi = NewGridIndexer()
	engine.gs = NewGridSearher(engine.gi)
	engine.gi.indexing()
}

func (engine *Engine) indexDocs(pts []gridData) {
	engine.gi.indexDocs(pts)
}

/*
	http://localhost:8888/index?pt=lo,la,id
	curl -XPUT http://localhost:8888/index -d pt=lo,la,id
*/
func (engine *Engine) index(w http.ResponseWriter, r *http.Request) {
	//Analytical parameters, the default is not resolved.
	r.ParseForm()
	fmt.Println(r.Form)
	for k, v := range r.Form {
		fmt.Fprintf(w, "key:%s\n", k)
		vs := strings.Split(v[0], ",")
		fmt.Fprintf(w, "val:%#v\n", vs)
		if len(vs) == 3 {
			tlo, err := strconv.Atoi(vs[0])
			if err != nil {
				continue
			}
			tla, err := strconv.Atoi(vs[1])
			if err != nil {
				continue
			}
			tid, err := strconv.Atoi(vs[2])
			if err != nil {
				continue
			}
			gd := gridData{int32(tlo), int32(tla), int32(tid)}
			engine.indexDocs([]gridData{gd})
		}
	}
}

/*
	http://localhost:8888/search
	http://localhost:8888/search?rect=left,top,right,bottom
	curl -XPUT http://localhost:8888/search -d rect=left,top,right,bottom
	http://localhost:8888/search?rect=9333748,3517838,9381410,3482092
	just for china.
*/
func (engine *Engine) search(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
L:
	for k, v := range r.Form {
		fmt.Fprintf(w, "key:%s\n", k)
		vs := strings.Split(v[0], ",")
		fmt.Fprintf(w, "val:%#v\n", vs)
		if len(vs) == 4 {
			tRect, ok := NewRectBy4String(vs)
			if !ok {
				tRect = genRandomRect()
			}
			fmt.Fprintf(w, "search rect: %d\n", tRect)
			startTime := time.Now()
			resIDs := engine.gs.search(tRect)
			st := time.Now().UnixNano() - startTime.UnixNano()
			searchTime := float64(st) / 1e6

			fmt.Fprintf(w, "search time: %f ms.\nsearch  res: %d\nres len: %d.",
				searchTime, resIDs, len(resIDs))
			break L
		}
	}
	if len(r.Form) == 0 {
		tRect := genRandomRect()
		fmt.Fprintf(w, "<random> search rect: %d\n", tRect)
		startTime := time.Now()
		resIDs := engine.gs.search(tRect)
		st := time.Now().UnixNano() - startTime.UnixNano()
		searchTime := float64(st) / 1e6

		fmt.Fprintf(w, "search time: %f ms.\nsearch  res: %d\nres len: %d.",
			searchTime, resIDs, len(resIDs))
	}
}

func (engine *Engine) close() {
	engine.gi.close()
}

func main() {
	engine := Engine{}
	engine.start()

	go func() {
		for {
			data := make([]gridData, 1000)
			for i := 0; i < 1000; i++ {
				data[i].lo = genRandomLo()
				data[i].la = genRandomLa()
				data[i].id = genRandomID()
			}
			engine.indexDocs(data)
		}
	}()

	http.HandleFunc("/search", engine.search)
	http.HandleFunc("/index", engine.index)
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
