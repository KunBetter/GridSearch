// Engine
package main

import (
	"encoding/json"
	"fmt"
	"github.com/KunBetter/GridSearch/stats"
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
	stats.StatsInit()
	OnInterrupt(func() {
		fmt.Println("start to stop Engine...")
		engine.close()
		fmt.Println("successfully stopped the engine.")
	})
}

func (engine *Engine) indexDocs(pts []gridData) {
	stats.IndexIn()
	engine.gi.indexDocs(pts)
}

func (engine *Engine) stats(w http.ResponseWriter, r *http.Request) {
	jm, _ := json.Marshal(stats.Stats())
	fmt.Fprintf(w, "%s", string(jm))
}

func (engine *Engine) mem(w http.ResponseWriter, r *http.Request) {
	jm, _ := json.Marshal(stats.MemStat())
	fmt.Fprintf(w, "%s", string(jm))
}

func (engine *Engine) disk(w http.ResponseWriter, r *http.Request) {
	ds := stats.NewDiskStatus("/Users/KunBetter")
	jm, _ := json.Marshal(ds)
	fmt.Fprintf(w, "%s", string(jm))
}

/*
	http://localhost:8888/index?pt=lo,la,id
	curl -XPUT http://localhost:8888/index -d pt=lo,la,id
*/
func (engine *Engine) index(w http.ResponseWriter, r *http.Request) {
	//Analytical parameters, the default is not resolved.
	r.ParseForm()
	for _, v := range r.Form {
		vs := strings.Split(v[0], ",")
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
	stats.SearchIn()
	r.ParseForm()
L:
	for _, v := range r.Form {
		vs := strings.Split(v[0], ",")
		if len(vs) == 4 {
			jmap := make(map[string]interface{})

			tRect, ok := NewRectBy4String(vs)
			if !ok {
				tRect = genRandomRect()
				jmap["type"] = "random"
			} else {
				jmap["type"] = "normal"
			}
			startTime := time.Now()
			resIDs := engine.gs.search(tRect)
			st := time.Now().UnixNano() - startTime.UnixNano()
			searchTime := float64(st) / 1e6

			jmap["rect"] = tRect
			jmap["took"] = searchTime
			jmap["unit"] = "ms"
			jmap["len"] = len(resIDs)
			jmap["hits"] = resIDs

			jm, _ := json.Marshal(jmap)
			fmt.Fprintf(w, "%s", string(jm))
			break L
		}
	}
	if len(r.Form) == 0 {
		jmap := make(map[string]interface{})

		tRect := genRandomRect()
		startTime := time.Now()
		resIDs := engine.gs.search(tRect)
		st := time.Now().UnixNano() - startTime.UnixNano()
		searchTime := float64(st) / 1e6

		jmap["type"] = "random"
		jmap["rect"] = tRect
		jmap["took"] = searchTime
		jmap["unit"] = "ms"
		jmap["len"] = len(resIDs)
		jmap["hits"] = resIDs

		jm, _ := json.Marshal(jmap)
		fmt.Fprintf(w, "%s", string(jm))
	}
}

func (engine *Engine) close() {
	engine.gi.close()
}

func (engine *Engine) Handler(r *http.ServeMux) {
	r.HandleFunc("/search", engine.search)
	r.HandleFunc("/index", engine.index)
	r.HandleFunc("/mem", engine.mem)
	r.HandleFunc("/disk", engine.disk)
	r.HandleFunc("/stats", engine.stats)
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

	r := http.NewServeMux()
	engine.Handler(r)

	err := http.ListenAndServe(":8888", r)
	if err != nil {
		panic(err)
	}
}
