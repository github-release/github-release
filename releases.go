package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/github-release/github-release/github"
	"github.com/dustin/go-humanize"
)

const (
	RELEASE_LIST_URI    = "/repos/%s/%s/releases"
	RELEASE_LATEST_URI  = "/repos/%s/%s/releases/latest"
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

func Releases(user, repo, authUser, token string) ([]Release, error) {
	var releases []Release
	client := github.NewClient(authUser, token, nil)
	client.SetBaseURL(EnvApiEndpoint)
	err := client.Get(fmt.Sprintf(RELEASE_LIST_URI, user, repo), &releases)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func latestReleaseApi(user, repo, authUser, token string) (*Release, error) {
	var release Release
	client := github.NewClient(authUser, token, nil)
	client.SetBaseURL(EnvApiEndpoint)
	return &release, client.Get(fmt.Sprintf(RELEASE_LATEST_URI, user, repo), &release)
}

func LatestRelease(user, repo, authUser, token string) (*Release, error) {
	// If latestReleaseApi DOESN'T give an error, return the release.
	if latestRelease, err := latestReleaseApi(user, repo, authUser, token); err == nil {
		return latestRelease, nil
	}

	// The enterprise api doesnt support the latest release endpoint. Get
	// all releases and compare the published date to get the latest.
	releases, err := Releases(user, repo, authUser, token)
	if err != nil {
		return nil, err
	}

	var latestRelIndex = -1
	maxDate := time.Time{}
	for i, release := range releases {
		if relDate := *release.Published; relDate.After(maxDate) {
			maxDate = relDate
			latestRelIndex = i
		}
	}
	if latestRelIndex == -1 {
		return nil, fmt.Errorf("could not find the latest release")
	}

	vprintln("Scanning ", len(releases), "releases, latest release is", releases[latestRelIndex])
	return &releases[latestRelIndex], nil
}

func ReleaseOfTag(user, repo, tag, authUser, token string) (*Release, error) {
	releases, err := Releases(user, repo, authUser, token)
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

func IncrementReleaseVersion(release *Release) *Release {
	// Handle nil release gracefully
	version := 0
	if release != nil {
		version, _ = strconv.Atoi(release.TagName[1:])
		version++
	} else {
		release = &Release{}
	}

	release.TagName = fmt.Sprintf("v%d", version)

	return release
}

/* find the release-id of the specified tag */
func IdOfTag(user, repo, tag, authUser, token string) (int, error) {
	release, err := ReleaseOfTag(user, repo, tag, authUser, token)
	if err != nil {
		return 0, err
	}

	return release.Id, nil
}
