package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/voxelbrain/goptions"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Options struct {
	Help      goptions.Help `goptions:"-h, --help, description='Show this help'"`
	Verbosity []bool        `goptions:"-v, --verbose, description='Be verbose'"`
	Quiet     bool          `goptions:"-q, --quiet, description='Do not print anything, even errors (except if --verbose is specified)'"`

	goptions.Verbs
	Upload struct {
		Token string   `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User  string   `goptions:"-u, --user, description='Github user (required if $GITHUB_USER not set)'"`
		Repo  string   `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag   string   `goptions:"-t, --tag, description='Git tag to upload for', obligatory"`
		Name  string   `goptions:"-n, --name, description='Name of the file', obligatory"`
		File  *os.File `goptions:"-f, --file, description='File to upload (use - for stdin)', rdonly, obligatory"`
	} `goptions:"upload"`
	Release struct {
		Token      string `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User       string `goptions:"-u, --user, description='Github user (required if $GITHUB_USER not set)'"`
		Repo       string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag        string `goptions:"-t, --tag, obligatory, description='Git tag to create a release from'"`
		Name       string `goptions:"-n, --name, description='Name of the release (defaults to tag)'"`
		Desc       string `goptions:"-d, --description, description='Description of the release (defaults to tag)'"`
		Draft      bool   `goptions:"--draft, description='The release is a draft'"`
		Prerelease bool   `goptions:"-p, --pre-release, description='The release is a pre-release'"`
	} `goptions:"release"`
	Delete struct {
		Token string `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User  string `goptions:"-u, --user, description='Github user (required if $GITHUB_USER not set)'"`
		Repo  string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag   string `goptions:"-t, --tag, obligatory, description='Git tag to create a release from'"`
	} `goptions:"delete"`
}

type Command func(Options) error

var commands = map[goptions.Verbs]Command{
	"upload":  uploadcmd,
	"release": releasecmd,
	"delete":  deletecmd,
}

var (
	VERBOSITY = 0
)

var (
	EnvToken string
	EnvUser  string
	EnvRepo  string
)

func init() {
	EnvToken = os.Getenv("GITHUB_TOKEN")
	EnvUser = os.Getenv("GITHUB_USER")
	EnvRepo = os.Getenv("GITHUB_REPO")
}

func main() {
	options := Options{}

	goptions.ParseAndFail(&options)

	if len(options.Verbs) == 0 {
		goptions.PrintHelp()
		return
	}

	VERBOSITY = len(options.Verbosity)

	if cmd, found := commands[options.Verbs]; found {
		err := cmd(options)
		if err != nil {
			if !options.Quiet {
				fmt.Println("error:", err)
			}
			os.Exit(1)
		}
	}
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
	resp, err := DoAuthRequest("POST", API_URL+uri, "application/json", token, reader)
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

	resp, err := httpDelete(API_URL+fmt.Sprintf("/repos/%s/%s/releases/%d",
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
