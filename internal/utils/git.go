package utils

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func CloneRepo(url string) error {
	fmt.Printf("Cloning repository: %s\n", url)

	_, err := git.PlainClone("./temp", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}
