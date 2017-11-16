// +build ignore

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"regexp"
)

var (
	startRE = regexp.MustCompile("(?m)^```go\n// examples/")
	end     = []byte("```")

	flagReadme = flag.String("readme", "README.md", "readme file")
	flagPath   = flag.String("path", "", "base path")
)

func main() {
	flag.Parse()

	buf, err := ioutil.ReadFile(*flagReadme)
	if err != nil {
		log.Fatal(err)
	}

	b := new(bytes.Buffer)
	for m := startRE.FindIndex(buf); m != nil; m = startRE.FindIndex(buf) {
		b.Write(buf[:m[0]])

		// grab filename
		st := m[0] + len("```go\n// ")
		en := bytes.IndexByte(buf[st:], '\n')
		filename := string(buf[st : st+en])

		// read example on disk
		b.WriteString("```go\n")
		ebuf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		b.Write(ebuf)
		b.WriteString("```\n")

		// chop buf
		buf = buf[m[1]:]
		end := bytes.Index(buf, end)
		if end == -1 {
			log.Fatal("improperly terminated example")
		}
		buf = buf[end+len("```\n"):]
	}
	b.Write(buf)

	err = ioutil.WriteFile(*flagReadme, b.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
