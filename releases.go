package main

import (
	"fmt"
	"strings"
)

const (
	RELEASE_LIST_URI = "/repos/%s/%s/releases"
)

type Release struct {
	Url       string  `json:"url"`
	UploadUrl string  `json:"upload_url"`
	Id        int     `json:"id"`
	Name      string  `json:"name"`
	TagName   string  `json:"tag_name"`
	Assets    []Asset `json:"assets"`
}

func (r *Release) CleanUploadUrl() string {
	bracket := strings.Index(r.UploadUrl, "{")

	if bracket == -1 {
		return r.UploadUrl
	}

	return r.UploadUrl[0:bracket]
}

type ReleaseCreate struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

func Releases(user, repo string) ([]Release, error) {
	var releases []Release
	err := GithubGet(fmt.Sprintf(RELEASE_LIST_URI, user, repo), &releases)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func ReleaseOfTag(user, repo, tag string) (*Release, error) {
	releases, err := Releases(user, repo)
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if release.TagName == tag {
			return &release, nil
		}
	}

	return nil, fmt.Errorf("could not find the release corresponding to tag %s", tag)
}

/* find the release-id of the specified tag */
func IdOfTag(user, repo, tag string) (int, error) {
	release, err := ReleaseOfTag(user, repo, tag)
	if err != nil {
		return 0, err
	}

	return release.Id, nil
}
