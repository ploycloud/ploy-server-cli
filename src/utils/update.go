package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ploycloud/ploy-server-cli/src/common"
)

var ReleaseEndpoint = "https://api.github.com/repos/ploycloud/ploy-server-cli/releases/latest"

type GitRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func CheckForUpdates() (string, bool, error) {
	res, err := http.Get(ReleaseEndpoint)

	if nil != err {
		return "", false, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if nil != err {
		return "", false, err
	}

	var release GitRelease
	if err := json.Unmarshal(body, &release); nil != err {
		return "", false, err
	}

	return release.TagName, release.TagName > common.CurrentCliVersion, nil
}

func SelfUpdate() (bool, error) {
	if 0 != os.Geteuid() {
		return false, fmt.Errorf("the update command must be run as root")
	}

	release, err := getLatestRelease()
	if nil != err {
		return false, err
	}

	fmt.Printf("Release: %s\n", release.TagName)

	assetURL := getAssetURL(release)
	if "" == assetURL {
		return false, fmt.Errorf("no suitable binary found for this system")
	}

	fmt.Printf("Download URL: %s\n", assetURL)

	// Download the new binary
	res, err := http.Get(assetURL)
	if nil != err {
		return false, err
	}

	defer res.Body.Close()

	// Create the new binary file
	tmpFile, err := os.CreateTemp("", "ploy-cli-update")
	if nil != err {
		return false, err
	}

	defer os.Remove(tmpFile.Name())

	// Write the body to the file
	_, err = io.Copy(tmpFile, res.Body)
	if nil != err {
		return false, err
	}

	// Close the file
	if err := tmpFile.Close(); nil != err {
		return false, err
	}

	// Make the file executable
	if err := os.Chmod(tmpFile.Name(), 0755); nil != err {
		return false, err
	}

	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		return false, err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false, err
	}

	// Rename the temporary file to the executable name
	return os.Rename(tmpFile.Name(), exe) == nil, nil
}

func getLatestRelease() (*GitRelease, error) {
	res, err := http.Get(ReleaseEndpoint)

	if nil != err {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if nil != err {
		return nil, err
	}

	var release GitRelease

	if err := json.Unmarshal(body, &release); nil != err {
		return nil, err
	}

	return &release, nil
}

func getAssetURL(release *GitRelease) string {
	arch := runtime.GOARCH
	systemOS := runtime.GOOS

	for _, asset := range release.Assets {
		if ".tar.gz" == filepath.Ext(asset.Name) &&
			"linux" == systemOS {
			if "amd64" == arch && "ploy-linux-amd64.tar.gz" == asset.Name {
				return asset.DownloadURL
			} else if "arm64" == arch && "ploy-linux-arm64.tar.gz" == asset.Name {
				return asset.DownloadURL
			}
		}
		// TODO: Add support for other OS and architectures
	}

	return ""
}
