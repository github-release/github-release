package main

import (
	"fmt"
	"os"

	"github.com/github-release/github-release/github"
	"github.com/voxelbrain/goptions"
)

const GH_URL = "https://github.com"

type Options struct {
	Help      goptions.Help `goptions:"-h, --help, description='Show this help'"`
	Verbosity []bool        `goptions:"-v, --verbose, description='Be verbose'"`
	Quiet     bool          `goptions:"-q, --quiet, description='Do not print anything, even errors (except if --verbose is specified)'"`
	Version   bool          `goptions:"--version, description='Print version'"`

	goptions.Verbs
	Download struct {
		Token    string `goptions:"-s, --security-token, description='Github token ($GITHUB_TOKEN if set). required if repo is private.'"`
		User     string `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		AuthUser string `goptions:"-a, --auth-user, description='Username for authenticating to the API (falls back to $GITHUB_AUTH_USER or $GITHUB_USER)'"`
		Repo     string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Latest   bool   `goptions:"-l, --latest, description='Download latest release (required if tag is not specified)',mutexgroup='input'"`
		Tag      string `goptions:"-t, --tag, description='Git tag to download from (required if latest is not specified)', mutexgroup='input',obligatory"`
		Name     string `goptions:"-n, --name, description='Name of the file', obligatory"`
	} `goptions:"download"`
	Upload struct {
		Token    string   `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User     string   `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		AuthUser string   `goptions:"-a, --auth-user, description='Username for authenticating to the API (falls back to $GITHUB_AUTH_USER or $GITHUB_USER)'"`
		Repo     string   `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag      string   `goptions:"-t, --tag, description='Git tag to upload to', obligatory"`
		Name     string   `goptions:"-n, --name, description='Name of the file', obligatory"`
		Label    string   `goptions:"-l, --label, description='Label (description) of the file'"`
		File     *os.File `goptions:"-f, --file, description='File to upload (use - for stdin)', rdonly, obligatory"`
		Replace  bool     `goptions:"-R, --replace, description='Replace asset with same name if it already exists (WARNING: not atomic, failure to upload will remove the original asset too)'"`
	} `goptions:"upload"`
	Release struct {
		Token                string `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User                 string `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		Repo                 string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag                  string `goptions:"-t, --tag, obligatory, description='Git tag to create a release from'"`
		Name                 string `goptions:"-n, --name, description='Name of the release (defaults to tag)'"`
		Desc                 string `goptions:"-d, --description, description='Release description, use - for reading a description from stdin (defaults to tag)'"`
		Target               string `goptions:"-c, --target, description='Commit SHA or branch to create release of (defaults to the repository default branch)'"`
		Draft                bool   `goptions:"--draft, description='The release is a draft'"`
		Prerelease           bool   `goptions:"-p, --pre-release, description='The release is a pre-release'"`
		GenerateReleaseNotes bool   `goptions:"-g, --generate-release-notes, description='Generate name and description if not given'"`
	} `goptions:"release"`
	Edit struct {
		Token      string `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User       string `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		AuthUser   string `goptions:"-a, --auth-user, description='Username for authenticating to the API (falls back to $GITHUB_AUTH_USER or $GITHUB_USER)'"`
		Repo       string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag        string `goptions:"-t, --tag, obligatory, description='Git tag to edit the release of'"`
		Name       string `goptions:"-n, --name, description='New name of the release (defaults to tag)'"`
		Desc       string `goptions:"-d, --description, description='New release description, use - for reading a description from stdin (defaults to tag)'"`
		Draft      bool   `goptions:"--draft, description='The release is a draft'"`
		Prerelease bool   `goptions:"-p, --pre-release, description='The release is a pre-release'"`
	} `goptions:"edit"`
	Delete struct {
		Token    string `goptions:"-s, --security-token, description='Github token (required if $GITHUB_TOKEN not set)'"`
		User     string `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		AuthUser string `goptions:"-a, --auth-user, description='Username for authenticating to the API (falls back to $GITHUB_AUTH_USER or $GITHUB_USER)'"`
		Repo     string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag      string `goptions:"-t, --tag, obligatory, description='Git tag of release to delete'"`
	} `goptions:"delete"`
	Info struct {
		Token    string `goptions:"-s, --security-token, description='Github token ($GITHUB_TOKEN if set). required if repo is private.'"`
		User     string `goptions:"-u, --user, description='Github repo user or organisation (required if $GITHUB_USER not set)'"`
		AuthUser string `goptions:"-a, --auth-user, description='Username for authenticating to the API (falls back to $GITHUB_AUTH_USER or $GITHUB_USER)'"`
		Repo     string `goptions:"-r, --repo, description='Github repo (required if $GITHUB_REPO not set)'"`
		Tag      string `goptions:"-t, --tag, description='Git tag to query (optional)'"`
		JSON     bool   `goptions:"-j, --json, description='Emit info as JSON instead of text'"`
	} `goptions:"info"`
}

type Command func(Options) error

var commands = map[goptions.Verbs]Command{
	"download": downloadcmd,
	"upload":   uploadcmd,
	"release":  releasecmd,
	"edit":     editcmd,
	"delete":   deletecmd,
	"info":     infocmd,
}

var (
	VERBOSITY = 0
)

var (
	// The user whose token is being used to authenticate to the API. If unset,
	// EnvUser is used.
	EnvAuthUser    string
	EnvToken       string
	EnvUser        string
	EnvRepo        string
	EnvApiEndpoint string
)

func init() {
	EnvToken = os.Getenv("GITHUB_TOKEN")
	EnvUser = os.Getenv("GITHUB_USER")
	EnvAuthUser = os.Getenv("GITHUB_AUTH_USER")
	EnvRepo = os.Getenv("GITHUB_REPO")
	EnvApiEndpoint = os.Getenv("GITHUB_API")

	if EnvAuthUser == "" {
		EnvAuthUser = EnvUser
	}
}

func main() {
	options := Options{}

	goptions.ParseAndFail(&options)

	if options.Version {
		fmt.Printf("github-release v%s\n", github.VERSION)
		return
	}

	if len(options.Verbs) == 0 {
		goptions.PrintHelp()
		return
	}

	VERBOSITY = len(options.Verbosity)
	github.VERBOSITY = VERBOSITY

	if cmd, found := commands[options.Verbs]; found {
		err := cmd(options)
		if err != nil {
			if !options.Quiet {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
			os.Exit(1)
		}
	}
}
