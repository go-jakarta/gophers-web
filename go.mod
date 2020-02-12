module gophers.id/gophers-web

require (
	cloud.google.com/go v0.52.0 // indirect
	github.com/NYTimes/gziphandler v1.1.1
	github.com/brankas/assetgen v0.0.0-20200202213138-f65a50a1f18a
	github.com/brankas/envcfg v0.1.0
	github.com/brankas/git-buildnumber v0.0.0-20200202205341-7c26e196a668
	github.com/brankas/goji v0.0.0-20191018231504-5bef0a1f9e4f
	github.com/brankas/netmux v0.1.1 // indirect
	github.com/brankas/sentinel v0.1.2 // indirect
	github.com/brankas/stringid v0.0.0-20191001010012-baeeb709f50a
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/digitalocean/godo v1.30.0 // indirect
	github.com/golang/gddo v0.0.0-20200203211524-5a4fa0264114
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/gorilla/csrf v1.6.2
	github.com/kenshaw/logrusmw v0.0.0-20180513035142-476e6892bf0a
	github.com/kenshaw/secure v0.0.0-20181217163002-d9facd3a9b63
	github.com/knq/ini v0.0.0-20191206014339-58b5e74713e0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/leonelquinteros/gotext v1.4.0
	github.com/miekg/dns v1.1.27 // indirect
	github.com/oschwald/maxminddb-golang v1.6.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0
	github.com/sirupsen/logrus v1.4.2
	github.com/tylerb/graceful v1.2.15
	github.com/valyala/quicktemplate v1.4.1
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678 // indirect
	google.golang.org/api v0.17.0 // indirect
	google.golang.org/genproto v0.0.0-20200211111953-2dc5924e3898 // indirect
	google.golang.org/grpc v1.27.1 // indirect
)

replace github.com/shurcooL/vfsgen => github.com/kenshaw/vfsgen v0.0.0-20181201224209-11cc086c1a6d

go 1.13
