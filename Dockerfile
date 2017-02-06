FROM scratch

ADD bin/docker/amd64/github-release /
CMD ["/github-release"]
