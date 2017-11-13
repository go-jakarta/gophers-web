// gophers-web is the web server for gophers.id.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
		// handle the go-get stuff
		if req.URL.Query().Get("go-get") == "1" {
			repo := strings.TrimPrefix(req.URL.Path, "/")
			if i := strings.Index(repo, "/"); i != -1 {
				repo = repo[:i]
			}

			// determine the repository
			repo = strings.TrimSuffix(repo, ".git")
			url := "https://github.com/go-jakarta/" + repo + ".git"
			if strings.ToLower(repo) == "gojakarta" {
				url = "https://github.com/go-jakarta/slides.git"
			}

			fmt.Fprintf(res, gitHTML, `gophers.id/`+repo+` git `+url, url, url)
			return
		}

		// do a redirect ...
		if strings.ToLower(strings.TrimPrefix(req.URL.Path, "/")) == "gojakarta" {
			http.Redirect(res, req, "https://meetup.com/GoJakarta", http.StatusMovedPermanently)
			return
		}

		fmt.Fprint(res, "nothing here\n")
	})

	listen := *flagHost + ":" + *flagPort
	log.Printf("listen: %s", listen)
	log.Fatal(http.ListenAndServe(listen, mux))
}

const (
	gitHTML = `<!DOCTYPE html>
<html>
<head>
  <meta name="go-import" content="%s">
</head>
<body>
  <a href="%s">%s</a>
</body>
</html>`
)
