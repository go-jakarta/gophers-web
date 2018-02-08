// +build ignore

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

var (
	flagReadme  = flag.String("readme", "README.md", "README file")
	flagType    = flag.String("type", "go", "file markdown type")
	flagBase    = flag.String("base", "examples/", "base path")
	flagComment = flag.String("comment", "//", "comment start")
)

func main() {
	flag.Parse()

	startRE := regexp.MustCompile("(?m)^```" + *flagType + "\r?\n" + *flagComment + " " + *flagBase)
	markdownEnd := []byte("```\n")

	// load file
	buf, err := ioutil.ReadFile(*flagReadme)
	if err != nil {
		log.Fatal(err)
	}

	// scan for start delimiter
	b := new(bytes.Buffer)
	for m := startRE.FindIndex(buf); m != nil; m = startRE.FindIndex(buf) {
		b.Write(buf[:m[0]])

		// grab filename
		st := m[0] + len("```"+*flagType+"\n"+*flagComment+" ")
		en := bytes.IndexByte(buf[st:], '\n')
		filename := strings.TrimSpace(string(buf[st : st+en]))

		// read example on disk
		b.WriteString("```" + *flagType + "\n")
		ebuf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		b.Write(ebuf)
		b.WriteString("```\n")

		// chop buf
		buf = buf[m[1]:]
		end := bytes.Index(buf, markdownEnd)
		if end == -1 {
			log.Fatal("improperly terminated markdown")
		}
		buf = buf[end+len(markdownEnd):]
	}
	b.Write(buf)

	err = ioutil.WriteFile(*flagReadme, b.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
