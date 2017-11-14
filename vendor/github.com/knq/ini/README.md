# About ini [![Build Status](https://travis-ci.org/knq/ini.svg)](https://travis-ci.org/knq/ini) [![Coverage Status](https://coveralls.io/repos/knq/ini/badge.svg?branch=master&service=github)](https://coveralls.io/github/knq/ini?branch=master) #

A simple [Go](http://www.golang.org/project/) package for manipulating
[ini files](https://en.wikipedia.org/wiki/INI_file).

This package is mostly a simple wrapper around the [parser package](/parser)
also in this repository. The parser package was implemented by generating a
[Pigeon](https://github.com/mna/pigeon/) parser from a
[PEG grammar](https://en.wikipedia.org/wiki/Parsing_expression_grammar).

This package can (with the correct configuration), correctly read [git
configuration](http://git-scm.com/docs/git-config) files, very simple
[TOML](https://github.com/toml-lang/toml) files, and [Java
Properties](https://en.wikipedia.org/wiki/.properties) files.

## Why Another ini File Package? ##

Prior to writing this package, a number of existing Go ini packages/parsers
were investigated. The available packages at the time did not have a complete
feature set needed, did not work well with badly formatted files, or their
parsers were not easily fixable.

As such, it was deemed necessary to write a package that could work with a
variety of badly formatted ini files, in idiomatic Go, and provide a simple
interface to reading/writing/manipulating ini files.

## Installation ##

Install the package via the following:

    go get -u github.com/knq/ini

## Usage ##

Please see [the GoDoc API page](http://godoc.org/github.com/knq/ini) for a full
API listing.

The ini package can be used similarly to the following:

```go
package main

import (
	"fmt"
	"log"

	"github.com/knq/ini"
)

var (
	data = `
	firstkey = one

	[some section]
	key = blah ; comment

	[another section]
	key = blah`

	gitconfig = `
	[difftool "gdmp"]
	cmd = ~/gdmp/x "$LOCAL" "$REMOTE"
	`
)

func main() {
	f, err := ini.LoadString(data)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	s := f.GetSection("some section")

	fmt.Printf("some section.key: %s\n", s.Get("key"))
	s.SetKey("key2", "another value")
	f.Write("out.ini")

	// create a gitconfig parser
	g, err := ini.LoadString(gitconfig)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	// setup gitconfig name/key manipulation functions
	g.SectionManipFunc = ini.GitSectionManipFunc
	g.SectionNameFunc = ini.GitSectionNameFunc

	fmt.Printf("difftool.gdmp.cmd: %s\n", g.GetKey("difftool.gdmp.cmd"))
}
```
