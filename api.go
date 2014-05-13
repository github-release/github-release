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
	GH_URL  = "https://github.com"
)

/* create a new request that sends the auth token */
func NewAuthRequest(method, url, bodyType, token string, headers map[string]string, body io.Reader) (*http.Request, error) {
	vprintln("creating request:", method, url, bodyType, token)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		switch v := body.(type) {
		case *os.File:
			/* when uploading a file, you need to either excplicitly set the
			 * Content-Length header or send a chunked request. Since the
			 * github upload server doesn't accept chunked encoding, we have
			 * to set the size of the file manually. If the file is actually
			 * stdin, this won't work, and we'll just quit for now.
			 *
			 * TODO: if stdin, read everything into a byte buffer, take the
			 * length and send it. */
			off, err := GetFileSize(v)
			if err != nil {
				return nil, err
			}

			req.ContentLength = off
			vprintln("setting content-length to", off)
		}
	}

	if bodyType != "" {
		req.Header.Set("Content-Type", bodyType)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func DoAuthRequest(method, url, bodyType, token string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := NewAuthRequest(method, url, bodyType, token, headers, body)
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
	resp, err := http.Get(ApiURL() + uri)
	if err != nil {
		return fmt.Errorf("could not fetch releases, %v", err)
	}
	defer resp.Body.Close()

	vprintln("GET", ApiURL()+uri, "->", resp)

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

func ApiURL() string {
	if "" == EnvApiEndpoint {
		return API_URL
	} else {
		return EnvApiEndpoint
	}
}
