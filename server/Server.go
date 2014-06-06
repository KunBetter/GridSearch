// Server
package main

import (
	"github.com/KunBetter/GridSearch"
	"net/http"
)

func main() {
	engine := GridSearch.Engine{}
	engine.Start()

	go func() {
		for {
			data := make([]GridSearch.GridData, 1000)
			for i := 0; i < 1000; i++ {
				data[i].LO = GridSearch.GenRandomLo()
				data[i].LA = GridSearch.GenRandomLa()
				data[i].ID = GridSearch.GenRandomID()
			}
			engine.IndexDocs(data)
		}
	}()

	r := http.NewServeMux()
	engine.Handler(r)

	err := http.ListenAndServe(":8888", r)
	if err != nil {
		panic(err)
	}
}
