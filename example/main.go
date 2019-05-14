//go:generate statiq -src=./public

package main

import (
	"log"
	"net/http"

	_ "github.com/bingoohuang/statiq/example/statiq"
	"github.com/bingoohuang/statiq/fs"
)

// Before buildling, run go generate.
// Then, run the main program and visit http://localhost:8080/public/hello.txt
func main() {
	statiqFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(statiqFS)))
	http.ListenAndServe(":8080", nil)
}
