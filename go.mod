module gophers.id/gophers-web

require (
	cloud.google.com/go v0.48.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1
	github.com/brankas/assetgen v0.0.0-20191018233711-b2695f03aa01
	github.com/brankas/envcfg v0.1.0
	github.com/brankas/git-buildnumber v0.0.0-20191018225650-9981460e62f9
	github.com/brankas/goji v0.0.0-20191018231504-5bef0a1f9e4f
	github.com/brankas/sentinel v0.1.1 // indirect
	github.com/brankas/stringid v0.0.0-20191001010012-baeeb709f50a
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/digitalocean/godo v1.25.0 // indirect
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gorilla/csrf v1.6.1
	github.com/kenshaw/logrusmw v0.0.0-20180513035142-476e6892bf0a
	github.com/kenshaw/secure v0.0.0-20181217163002-d9facd3a9b63
	github.com/knq/ini v0.0.0-20191109065004-cbd1d95dcaf6 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/miekg/dns v1.1.22 // indirect
	github.com/oschwald/maxminddb-golang v1.5.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0
	github.com/sirupsen/logrus v1.4.2
	github.com/tylerb/graceful v1.2.15
	github.com/valyala/quicktemplate v1.4.1
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708 // indirect
	golang.org/x/net v0.0.0-20191112182307-2180aed22343 // indirect
	golang.org/x/sys v0.0.0-20191113165036-4c7a9d0fe056 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.1 // indirect
)

replace github.com/shurcooL/vfsgen => github.com/kenshaw/vfsgen v0.0.0-20181201224209-11cc086c1a6d

go 1.13
