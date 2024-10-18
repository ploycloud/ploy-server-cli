package github

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	rawGitHubURL = "https://raw.githubusercontent.com/ploycloud/ploy-server-cli/main/src/assets/docker/"
)

func GetDockerComposeTemplate(filename string) ([]byte, error) {
	url := rawGitHubURL + filename
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template from GitHub: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch template from GitHub: status code %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
