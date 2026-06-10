package git

import (
	"os"
	"os/exec"
	"path/filepath"
)

// first make the files, then git init
func InitRepo(ignore Template, license Template) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// write to readme file
	if err := os.WriteFile(filepath.Join(cwd, "README.md"), []byte(ReadmeTemplate), 0644); err != nil {
		return err
	}

	// write to gitignore (if selected)
	if ignore.Content != "" {
		if err := os.WriteFile(filepath.Join(cwd, ".gitignore"), []byte(ignore.Content), 0644); err != nil {
			return err
		}
	}

	// write to license (if selected)
	if license.Content != "" {
		if err := os.WriteFile(filepath.Join(cwd, "LICENSE"), []byte(license.Content), 0644); err != nil {
			return err
		}
	}

	// finally, make the git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = cwd
	return cmd.Run()
}
