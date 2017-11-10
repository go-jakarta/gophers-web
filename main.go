// gophers-web is the web server for gophers.id.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	flagListen = flag.String("listen", os.Getenv("HOST")+":"+os.Getenv("PORT"), "listen")
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "nothing here.\n")
	})
	log.Fatal(http.ListenAndServe(*flagListen, mux))
}
