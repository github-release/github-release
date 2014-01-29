package main

import (
	"fmt"
)

const (
	TAGS_URI = "/repos/%s/%s/tags"
)

type Tag struct {
	Name       string `json:"name"`
	Commit     Commit `json:"commit"`
	ZipBallUrl string `json:"zipball_url"`
	TarBallUrl string `json:"tarball_url"`
}

/* get the tags associated with a repo */
func Tags(user, repo string) ([]Tag, error) {
	var tags []Tag
	err := GithubGet(fmt.Sprintf(TAGS_URI, user, repo), &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
