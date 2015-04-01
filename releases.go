package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	RELEASE_LIST_URI    = "/repos/%s/%s/releases%s"
	RELEASE_LATEST_URI    = "/repos/%s/%s/releases/latest%s"
	RELEASE_DATE_FORMAT = "02/01/2006 at 15:04"
)

type Release struct {
	Url         string     `json:"url"`
	PageUrl     string     `json:"html_url"`
	UploadUrl   string     `json:"upload_url"`
	Id          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"body"`
	TagName     string     `json:"tag_name"`
	Draft       bool       `json:"draft"`
	Prerelease  bool       `json:"prerelease"`
	Created     *time.Time `json:"created_at"`
	Published   *time.Time `json:"published_at"`
	Assets      []Asset    `json:"assets"`
}

func (r *Release) CleanUploadUrl() string {
	bracket := strings.Index(r.UploadUrl, "{")

	if bracket == -1 {
		return r.UploadUrl
	}

	return r.UploadUrl[0:bracket]
}

func (r *Release) String() string {
	str := make([]string, len(r.Assets)+1)
	str[0] = fmt.Sprintf(
		"%s, name: '%s', description: '%s', id: %d, tagged: %s, published: %s, draft: %v, prerelease: %v",
		r.TagName, r.Name, r.Description, r.Id,
		timeFmtOr(r.Created, RELEASE_DATE_FORMAT, ""),
		timeFmtOr(r.Published, RELEASE_DATE_FORMAT, ""),
		Mark(r.Draft), Mark(r.Prerelease))

	for idx, asset := range r.Assets {
		str[idx+1] = fmt.Sprintf("  - artifact: %s, downloads: %d, state: %s, type: %s, size: %s, id: %d",
			asset.Name, asset.Downloads, asset.State, asset.ContentType,
			humanize.Bytes(asset.Size), asset.Id)
	}

	return strings.Join(str, "\n")
}

type ReleaseCreate struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

func Releases(user, repo, token string) ([]Release, error) {
	if token != "" {
		token = "?access_token=" + token
	}
	var releases []Release
	err := GithubGet(fmt.Sprintf(RELEASE_LIST_URI, user, repo, token), &releases)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func LatestReleaseApi(user, repo, token string) (*Release, error) {
	if token != "" {
		token = "?access_token=" + token
	}
	var release Release
	err := GithubGet(fmt.Sprintf(RELEASE_LATEST_URI, user, repo, token), &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func LatestRelease(user, repo, token string) (*Release, error) {
	var latestRelease *Release
	latestRelease, err := LatestReleaseApi(user, repo, token)

	// enterprise api doesnt support the latest release endpoint
	// get all releases and see published date to get the latest

	if err != nil {
		releases, err := Releases(user, repo, token)
		if err != nil {
			return nil, err
		}
		var latestRelIndex = -1
		var maxDate time.Time = time.Time{}
		for i, release := range releases {
			var relDate = *(release.Published)
			if relDate.After(maxDate) {
				maxDate = relDate
				latestRelIndex = i
			}
		}
		if(latestRelIndex!=-1) {
			latestRelease = &releases[latestRelIndex]
		}
	}
	vprintln("latest release is ->",latestRelease)
	return latestRelease, nil
}

func ReleaseOfTag(user, repo, tag, token string) (*Release, error) {
	releases, err := Releases(user, repo, token)
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
func IdOfTag(user, repo, tag, token string) (int, error) {
	release, err := ReleaseOfTag(user, repo, tag, token)
	if err != nil {
		return 0, err
	}

	return release.Id, nil
}
