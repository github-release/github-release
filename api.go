package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	API_URL = "https://api.github.com"
)

/* create a new request that sends the auth token */
func NewAuthRequest(method, url, bodyType, token string, body io.Reader) (*http.Request, error) {
	vprintln("creating request:", method, url, bodyType, token)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		switch v := body.(type) {
		case *os.File:
			/* apparently chunking doesnt work yet...
			 * vprintln("OS.FILE detected, chunking!", v)
			 * req.TransferEncoding = []string{"chunked"} */

			/* then we explicitly read the file... (let's hope it's not stdin) */
			off, err := GetFileSize(v)
			if err != nil {
				return nil, err
			}

			req.ContentLength = off
			vprintln("setting content-length to", off)
		}
	}

	req.Header.Set("Content-Type", bodyType)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	return req, nil
}

func DoAuthRequest(method, url, bodyType, token string, body io.Reader) (*http.Response, error) {
	req, err := NewAuthRequest(method, url, bodyType, token, body)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GithubGet(uri string, v interface{}) error {
	resp, err := http.Get(API_URL + uri)
	if err != nil {
		return fmt.Errorf("could not fetch releases, %v", err)
	}
	defer resp.Body.Close()

	vprintln("GET", API_URL+uri, "->", resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github did not response with 200 OK but with %v", resp.Status)
	}

	if VERBOSITY == 0 {
		if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("could not unmarshall JSON into Release struct, %v", err)
		}
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		vprintln("BODY", string(body))

		if err = json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("could not unmarshall JSON into Release struct, %v", err)
		}
	}

	return nil
}
