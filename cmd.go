package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func infocmd(opt Options) error {
	user := nvls(opt.Info.User, EnvUser)
	repo := nvls(opt.Info.Repo, EnvRepo)
	tag := opt.Upload.Tag

	if user == "" || repo == "" {
		return fmt.Errorf("user and repo need to be passed as arguments")
	}

	/* list all tags */
	tags, err := Tags(user, repo)
	if err != nil {
		return fmt.Errorf("could not fetch tags, %v", err)
	}

	fmt.Println("git tags:")
	for _, tag := range tags {
		fmt.Println("-", tag.String())
	}

	/* list releases + assets */
	fmt.Println("releases:")
	var releases []Release
	if tag == "" {
		/* get all releases */
		vprintf("%v/%v: getting information for all releases\n", user, repo)
		releases, err = Releases(user, repo)
		if err != nil {
			return err
		}
	} else {
		/* get only one release */
		vprintf("%v/%v/%v: getting information for the release\n", user, repo, tag)
		release, err := ReleaseOfTag(user, repo, tag)
		if err != nil {
			return err
		}
		releases = []Release{*release}
	}

	for _, release := range releases {
		fmt.Println("-", release.String())
	}

	return nil
}

func uploadcmd(opt Options) error {
	user := nvls(opt.Upload.User, EnvUser)
	repo := nvls(opt.Upload.Repo, EnvRepo)
	token := nvls(opt.Upload.Token, EnvToken)
	tag := opt.Upload.Tag
	name := opt.Upload.Name
	file := opt.Upload.File

	vprintln("uploading...")

	if file == nil {
		return fmt.Errorf("provided file was not valid")
	}

	if err := ValidateCredentials(user, repo, token, tag); err != nil {
		return err
	}

	/* find the release corresponding to the entered tag, if any */
	rel, err := ReleaseOfTag(user, repo, tag)
	if err != nil {
		return err
	}

	v := url.Values{}
	v.Set("name", name)

	url := rel.CleanUploadUrl() + "?" + v.Encode()

	resp, err := DoAuthRequest("POST", url, "application/octet-stream",
		token, file)
	if err != nil {
		return fmt.Errorf("can't create upload request to %v, %v", url, err)
	}
	defer resp.Body.Close()

	vprintln("RESPONSE:", resp)
	if resp.StatusCode != http.StatusCreated {
		if msg, err := ToMessage(resp.Body); err == nil {
			return fmt.Errorf("could not upload, status code (%v), %v",
				resp.Status, msg)
		} else {
			return fmt.Errorf("could not upload, status code (%v)", resp.Status)
		}
	}

	if VERBOSITY != 0 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error while reading response, %v", err)
		}
		vprintln("BODY:", string(body))
	}

	return nil
}

func ValidateCredentials(user, repo, token, tag string) error {
	if user == "" {
		return fmt.Errorf("empty user")
	}
	if repo == "" {
		return fmt.Errorf("empty repo")
	}
	if tag == "" {
		return fmt.Errorf("empty tag")
	}
	if token == "" {
		return fmt.Errorf("empty token")
	}
	return nil
}

func releasecmd(opt Options) error {
	user := nvls(opt.Release.User, EnvUser)
	repo := nvls(opt.Release.Repo, EnvRepo)
	token := nvls(opt.Release.Token, EnvToken)
	tag := opt.Release.Tag
	name := nvls(opt.Release.Name, tag)
	desc := nvls(opt.Release.Desc, tag)
	draft := opt.Release.Draft
	prerelease := opt.Release.Prerelease

	vprintln("releasing...")

	if err := ValidateCredentials(user, repo, token, tag); err != nil {
		return err
	}

	params := ReleaseCreate{
		TagName:    tag,
		Name:       name,
		Body:       desc,
		Draft:      draft,
		Prerelease: prerelease,
	}

	/* when verbosity is off, we don't need the body and we can create a
	 * version with lower overhead (I actually notice this to be slower,
	 * should benchmark...) */
	var (
		reader  io.Reader
		rawjson string
	)
	if VERBOSITY == 0 {
		var writer io.WriteCloser
		reader, writer = io.Pipe()
		enc := json.NewEncoder(writer)
		go func() {
			err := enc.Encode(params)
			defer writer.Close()
			if err != nil {
				/* TODO: we can probably end cleaner here... */
				panic(fmt.Errorf("can't encode release creation params, %v", err))
			}
		}()
	} else {
		jsonBytes, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("can't encode release creation params, %v", err)
		}
		rawjson = string(jsonBytes)
		reader = bytes.NewReader(jsonBytes)
	}

	uri := fmt.Sprintf("/repos/%s/%s/releases", user, repo)
	resp, err := DoAuthRequest("POST", ApiURL()+uri, "application/json", token, reader)
	if err != nil {
		return fmt.Errorf("while submitting %v, %v", rawjson, err)
	}
	defer resp.Body.Close()

	vprintln("RESPONSE:", resp)
	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == 422 {
			return fmt.Errorf("github returned %v (this is probably because the release already exists)",
				resp.Status)
		}
		return fmt.Errorf("github returned %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response, %v", err)
	}
	vprintln("BODY:", string(body))

	return nil
}

func editcmd(opt Options) error {
	cmdopt := opt.Edit
	user := nvls(cmdopt.User, EnvUser)
	repo := nvls(cmdopt.Repo, EnvRepo)
	token := nvls(cmdopt.Token, EnvToken)
	tag := cmdopt.Tag
	name := nvls(cmdopt.Name, tag)
	desc := nvls(cmdopt.Desc, tag)
	draft := cmdopt.Draft
	prerelease := cmdopt.Prerelease

	vprintln("editing...")

	if err := ValidateCredentials(user, repo, token, tag); err != nil {
		return err
	}

	id, err := IdOfTag(user, repo, tag)
	if err != nil {
		return err
	}

	vprintf("release %v has id %v\n", tag, id)

	/* the release create struct works for editing releases as well */
	params := ReleaseCreate{
		TagName:    tag,
		Name:       name,
		Body:       desc,
		Draft:      draft,
		Prerelease: prerelease,
	}

	/* encode the parameters as JSON, as required by the github API */
	payload, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("can't encode release creation params, %v", err)
	}

	uri := fmt.Sprintf("/repos/%s/%s/releases/%d", user, repo, id)
	resp, err := DoAuthRequest("PATCH", ApiURL()+uri, "application/json",
		token, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("while submitting %v, %v", string(payload), err)
	}
	defer resp.Body.Close()

	vprintln("RESPONSE:", resp)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 422 {
			return fmt.Errorf("github returned %v (this is probably because the release already exists)",
				resp.Status)
		}
		return fmt.Errorf("github returned unexpected status code %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response, %v", err)
	}
	vprintln("BODY:", string(body))

	return nil
}

func deletecmd(opt Options) error {
	user, repo, token, tag := nvls(opt.Delete.User, EnvUser),
		nvls(opt.Delete.Repo, EnvRepo),
		nvls(opt.Delete.Token, EnvToken),
		opt.Delete.Tag
	vprintln("deleting...")

	id, err := IdOfTag(user, repo, tag)
	if err != nil {
		return err
	}

	vprintf("release %v has id %v\n", tag, id)

	resp, err := httpDelete(ApiURL()+fmt.Sprintf("/repos/%s/%s/releases/%d",
		user, repo, id), token)
	if err != nil {
		return fmt.Errorf("release deletion unsuccesful, %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("could not delete the release corresponding to tag %s on repo %s/%s",
			tag, user, repo)
	}

	return nil
}

func httpDelete(url, token string) (*http.Response, error) {
	resp, err := DoAuthRequest("DELETE", url, "application/json", token, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
