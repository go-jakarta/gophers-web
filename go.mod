module gophers.id/gophers-web

require (
	cloud.google.com/go v0.46.2 // indirect
	github.com/NYTimes/gziphandler v1.1.1
	github.com/brankas/assetgen v0.0.0-20190914232448-f41838b3b9ba
	github.com/brankas/envcfg v0.1.0
	github.com/brankas/git-buildnumber v0.0.0-20190405072123-d52ef46b19af
	github.com/brankas/sentinel v0.1.1 // indirect
	github.com/brankas/stringid v0.0.0-20180515093455-3d3c553d8e97
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/digitalocean/godo v1.20.0 // indirect
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/csrf v1.6.1
	github.com/kenshaw/gojiid v0.0.0-20170710044130-982f5f15c83b
	github.com/kenshaw/logrusmw v0.0.0-20180513035142-476e6892bf0a
	github.com/kenshaw/secure v0.0.0-20181217163002-d9facd3a9b63
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/miekg/dns v1.1.17 // indirect
	github.com/oschwald/maxminddb-golang v1.5.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/tylerb/graceful v1.2.15
	github.com/valyala/quicktemplate v1.2.0
	go.opencensus.io v0.22.1 // indirect
	goji.io v2.0.2+incompatible
	golang.org/x/sys v0.0.0-20190913121621-c3b328c6e5a7 // indirect
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/grpc v1.23.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
)

replace github.com/shurcooL/vfsgen => github.com/kenshaw/vfsgen v0.0.0-20181201224209-11cc086c1a6d

go 1.13
