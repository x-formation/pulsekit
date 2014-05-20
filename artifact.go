package pulse

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// ArtifactFetcher is type for fetching artifacts based on info from BuildArtifact type
type ArtifactFetcher struct {
	Client        *http.Client
	tok, dir, url string
}

var errEmptyRes = errors.New("fakerpc: empty response")

// NewArtifactFetcher returns new ArtifactFetcher
func NewArtifactFetcher(url, tok, dir string) *ArtifactFetcher {
	return &ArtifactFetcher{Client: &http.Client{}, tok: tok, dir: dir, url: url}
}

// Fetch prepares file paths to save artifact files and calls downloadFile
func (af *ArtifactFetcher) Fetch(a *BuildArtifact, project string) (err error) {
	var link string
	if link, err = url.QueryUnescape(a.Permalink); err != nil {
		return err
	}
	basepath := filepath.Join(af.dir, project, a.Stage, a.Command, a.Name)
	urls := af.buildURLs(link, a.Files)
	for i := range a.Files {
		path := filepath.Join(basepath, filepath.Dir(a.Files[i]))

		if _, err = os.Stat(path); os.IsNotExist(err) {
			if err = os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}
		if err = af.fetchSingle(filepath.Join(path, filepath.Base(a.Files[i])), urls[i]); err != nil {
			return err
		}
	}
	return nil
}

// buildURLs builds URLs to download files within artifact
func (af *ArtifactFetcher) buildURLs(link string, files []string) []string {
	links := make([]string, len(files))
	for i := range files {
		links[i] = af.url + filepath.Join(link, files[i])
	}
	return links
}

// fetchSingle downloads file from url and saves it as filename
func (af *ArtifactFetcher) fetchSingle(filename string, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("PULSE_API_TOKEN", af.tok)
	resp, err := af.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.ContentLength == 0 {
		return errEmptyRes
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	return nil
}
