//go:generate statical -src=./public

package main

import (
	"log"
	"net/http"

	_ "github.com/bingoohuang/statical/example/statical"
	"github.com/bingoohuang/statical/fs"
)

// Before buildling, run go generate.
// Then, run the main program and visit http://localhost:8080/public/hello.txt
func main() {
	staticalFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(staticalFS)))
	http.ListenAndServe(":8080", nil)
}
