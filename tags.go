package main

import (
	"fmt"

	"github.com/mistsys/github-release/github"
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

func (t *Tag) String() string {
	return t.Name + " (commit: " + t.Commit.Url + ")"
}

// Get the tags associated with a repo.
func Tags(user, repo, authUser, token string) ([]Tag, error) {
	var tags []Tag
	client := github.NewClient(authUser, token, nil)
	client.SetBaseURL(EnvApiEndpoint)
	return tags, client.Get(fmt.Sprintf(TAGS_URI, user, repo), &tags)
}

func TagOfTag(user, repo, tag, authUser, token string) (*Tag, error) {
	var tags []Tag
	client := github.NewClient(authUser, token, nil)
	client.SetBaseURL(EnvApiEndpoint)
	url := fmt.Sprintf(TAGS_URI, user, repo)

	for url != "" {
		next, err := client.GetPaginated(url, 1, &tags)
		if err != nil {
			return nil, err
		}

		for _, t := range tags {
			if t.Name == tag {
				return &t, nil
			}
		}

		url = next
	}

	return nil, fmt.Errorf("could not find the tag corresponding to tag %s", tag)
}
