package main

import (
	"github.com/cristaloleg/flache"
	"net/http"
)

var cache flache.Cacher

func main() {
	http.HandleFunc("/", handleHTTPConnection)
	http.ListenAndServe(":8080", nil)
}

func handleHTTPConnection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		key := r.Form.Get("key")
		if key == "" {
			return
		}
		value, _, ok := cache.GetExt(key)
		if !ok {
			return
		}
		if _, ok2 := value.(string); ok2 {
			w.Write(value.([]byte))
		} else if _, ok2 := value.([]byte); ok2 {
			w.Write(value.([]byte))
		}

	case "PUT":
		key := r.Form.Get("key")
		if key == "" {
			return
		}
		value := r.Form.Get("value")
		if value == "" {
			return
		}
		cache.Add(key, value)

	case "HEAD":
		key := r.Form.Get("key")
		if key == "" {
			return
		}
		exists, alive := cache.Check(key)

		res := []byte{0, 0}
		if exists {
			res[0] = 1
		}
		if alive {
			res[1] = 1
		}
		w.Write(res)
	}
}
