package main

import (
	"log"
	"net/http"

	z01 "z01/functions"
)

func main() {
	if err := z01.CheckBannerChanges(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", z01.HomeHandler)
	http.HandleFunc("/ascii-art", z01.AsciiArtHandler)
	
	staticFiles := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	http.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		for _, header := range []string{"If-Match", "If-None-Match", "If-Modified-Since", "If-Unmodified-Since", "If-Range"} {
			r.Header.Del(header)
		}
		staticFiles.ServeHTTP(w, r)
	}))

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
