# About gophers-web

The website that runs [gophers.id](https://gophers.id/), the
[GoJakarta](https://meetup.com/GoJakarta) homepage.

## Installing and Running

```sh
# retrieve the code
$ go get -u -d gophers.id/gophers-web

# change to the working path
$ cd $GOPATH/src/gophers.id/gophers-web

# copy environment config
$ cp env/sample.config env/config

# install assetgen (used for generating static assets)
$ go get -u github.com/brankas/assetgen

# build executable and run
$ go generate && go build && ./gophers-web
```

## Updating Templates + Translation Messages

```sh
# change to the working path
$ cd $GOPATH/src/gophers.id/gophers-web

# edit templates
$ vi assets/templates/*.html

# update messages
$ ./misc/update-messages.sh

# translate messages
$ poedit assets/locales/id.po

# add locales/
$ git add assets/locales/id.po assets/templates/*.html

# commit and push
$ git commit -m 'Changing content' && git push
```

## Deploying

This app is deployed on the web as a Dokku app.

```sh
# add a dokku remote
git remote add dokku dokku@dokku01.sgp1.gophers.id:gophers-web

# push
git push dokku master
```

## Notes

This is not a static generated website (ie, Hugo), because it also serves as a
way to demonstrate writing small web sites with Go, as well as providing some
other slightly more advanced features that cannot be accomplished with a static
website alone.
