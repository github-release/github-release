github-release
==============

A small commandline app written in Go that allows you to easily create
and delete releases of your projects on Github. In addition it allows
you to attach files to those releases.

It interacts with the [github releases API](http://developer.github.com/v3/repos/releases).
Though it's entirely possibly to [do all these things with
cURL](https://github.com/blog/1645-releases-api-preview), It's not
really that user-friendly. For example, you need to first query the API
to find the id of the release you want, before you can upload an
artifact. `github-release` takes care of those little details.

How to install
==============

If you already have Go, you can just do:

```sh
go get github.com/aktau/github-release
```

This will automatically download, compile and install the app.

After that you should have a `github-release` executable in your
`$GOPATH/bin`.

How to use
==========

**NOTE**: for these examples I've [created a github
token](https://help.github.com/articles/creating-an-access-token-for-command-line-use)
and set in as the environment variable `$GITHUB_TOKEN`.

```sh
# set your token
export GITHUB_TOKEN=...

# check the help
$ github-release --help

# make your tag and upload
$ git tag ... && git push --tags

# check the current tags and existing releases of the repo
$ github-release info -u aktau -r gofinance
git tags:
- v0.1.0 (commit: https://api.github.com/repos/aktau/gofinance/commits/f562727ce83ce8971a8569a1879219e41d56a756)
releases:
- v0.1.0, name: 'hoary ungar', description: 'something something dark side 2', id: 166740, tagged: 29/01/2014 at 14:27, published: 30/01/2014 at 16:20, draft: ✔, prerelease: ✗
  - artifact: github.go, downloads: 0, state: uploaded, type: application/octet-stream, size: 1.9KB, id: 68616

# create a formal release
$ github-release release \
    --user aktau \
    --repo gofinance \
    --tag v0.1.0 \
    --name "the wolf of source street" \
    --description "Not a movie, contrary to popular opinion. Still, my first release!" \
    --pre-release

# upload a file, for example the OSX/AMD64 binary of my gofinance app
$ github-release upload \
    --user aktau \
    --repo gofinance \
    --tag v0.1.0 \
    --name "gofinance-osx-amd64" \
    --file bin/darwin/amd64/gofinance

# upload other files...
$ github-release upload ...

# you're not happy with it, so delete it
$ github-release delete \
    --user aktau \
    --repo gofinance \
    --tag v0.1.0
```

Used libraries
==============

| Package | Description | License |
| --- | --- | --- |
| [github.com/dustin/go-humanize](https://github.com/dustin/go-humanize) | humanize file sizes | MIT |

Copyright
=========

Copyright (c) 2014, Nicolas Hillegeer. All rights reserved.
