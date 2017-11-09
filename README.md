# About gophers-web

The website that runs [gophers.id](https://gophers.id/), the
[GoJakarta](https://meetup.com/GoJakarta) homepage.

## Installing and Running

```sh
$ go get -u -d gophers.id/gophers-web
$ cd $GOPATH/src/gophers.id/gophers-web
$ go generate
$ go build
$ ./gophers-web
```

## Deploying

This app is deployed on the web as a Dokku app.

```sh
# add a dokku remote
git remote add dokku dokku@dokku01-sgp1.gophers.id

# push
git push dokku master
```

## Notes

This is not a static generated website (ie, Hugo), because it also serves as a
way to demonstrate writing small web sites with Go, as well as providing some
other slightly more advanced features that cannot be accomplished with a static
website alone.
