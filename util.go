package main

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

/* oracle nvl, return first non-empty string */
func nvls(xs ...string) string {
	for _, s := range xs {
		if s != "" {
			return s
		}
	}

	return ""
}

func vprintln(a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Println(a...)
	}

	return 0, nil
}

func vprintf(format string, a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}

// formats time `t` as `fmt` if it is not nil, otherwise returns `def`
func timeFmtOr(t *time.Time, fmt, def string) string {
	if t == nil {
		return def
	}
	return t.Format(fmt)
}

func gitUserAndRepo() (string, string) {
	var repo string
	var user string
	origin, execErr := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if execErr == nil {
		originStr := string(origin[:])
		if strings.Contains(originStr, "git@") {
			originSegments := strings.Split(originStr, ":")
			uriSegments := strings.Split(originSegments[1], "/")
			repoSegments := strings.Split(uriSegments[1], ".")
			repo = repoSegments[0]
			user = uriSegments[0]
		} else {
			url, parseErr := url.Parse(originStr)
			if parseErr == nil {
				uriSegments := strings.Split(url.Path, "/")
				repoSegments := strings.Split(uriSegments[2], ".")
				repo = repoSegments[0]
				user = uriSegments[1]
			}
		}
	}
	return user, repo
}
