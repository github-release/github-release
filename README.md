github-release
==============

A small commandline app written in Go that allows you to easily create
and delete releases of your projects on Github. In addition it allows
you to attach files to those releases.

How to install (if you already have Go)
==============

```sh
go get github.com/aktau/github-release
```

After that you should have a `github-release` executable in your
`$GOPATH/bin`.

How to use
==========

```sh
# check the help
$ github-release --help

# make your tag and upload
$ git tag ... && git push --tags

# create a formal release
$ github-release release --user aktau \
    --repo gofinance \
    --tag v0.1.0 \
    --name "the wolf of bug street" \
    --description "Not a movie, contrary to popular opinion. Still, my first release!" \
    --pre-release

# upload a file, for example the OSX/AMD64 binary of my gofinance app
$ github-release upload --user aktau \
    --repo gofinance \
    --tag v0.1.0 \
    --name "gofinance-osx-amd64" \
    --file bin/darwin/amd64/gofinance

# upload other files...
$ github-release upload ...

# you're not happy with it, so delete it
$ github-release delete --user aktau \
    --repo gofinance \
    --tag v0.1.0
```

Copyright
=========

Copyright (c) 2014, Nicolas Hillegeer. All rights reserved.
