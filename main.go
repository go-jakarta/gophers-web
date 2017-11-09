// gophers-web is the web server for gophers.id.
package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	flagListen = flag.String("listen", "localhost:3000", "listen")
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "nothing here.\n")
	})
	http.ListenAndServe(*flagListen, mux)
}
