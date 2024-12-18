package docker

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var getGitHubURL = func() string {
	return "https://raw.githubusercontent.com/ploycloud/ploy-server-cli/main/docker/"
}

func GetDockerComposeTemplate(filename string) ([]byte, error) {
	url := getGitHubURL() + filename
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

// Docker Compose template references
const (
	WPComposeStaticTemplate  = "wp/wp-compose-static.yml"
	WPComposeDynamicTemplate = "wp/wp-compose-dynamic.yml"
	MySQLComposeTemplate     = "databases/mysql-compose.yml"
)
