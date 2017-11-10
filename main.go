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
	flagHost = flag.String("host", os.Getenv("HOST"), "host")
	flagPort = flag.String("port", os.Getenv("PORT"), "port")
)

func main() {
	flag.Parse()

	if *flagPort == "" {
		*flagPort = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, "one more test\n")
	})

	listen := *flagHost + ":" + *flagPort
	log.Printf("listen: %s", listen)
	log.Fatal(http.ListenAndServe(listen, mux))
}
