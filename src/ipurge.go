package ipurge

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	APP_HOME_PATH = filepath.Join(os.Getenv("HOME"), "Library")
	APP_ROOT_PATH = filepath.Join("/", "Applications")
	LIB_ROOT_PATH = filepath.Join("/", "Library")
	PRIV_VAR_PATH = filepath.Join("/", "private", "var")
	USR_LOCL_PATH = filepath.Join("/", "usr", "local")
	PATHS         = make(map[string]int) // Actual path (string) and MaxDepth (int)
)

func init() {

	PATHS[APP_HOME_PATH] = 1
	PATHS[APP_ROOT_PATH] = 1
	PATHS[LIB_ROOT_PATH] = 1
	PATHS[USR_LOCL_PATH] = 1
	PATHS[PRIV_VAR_PATH] = 2

	if s, err := getConfCMD("DARWIN_USER_CACHE_DIR"); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	} else {
		PATHS[filepath.Dir(s)] = 1
	}

	if s, err := getConfCMD("DARWIN_USER_TEMP_DIR"); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	} else {
		PATHS[filepath.Dir(s)] = 1
	}
}

func getConfCMD(s string) (string, error) {
	out, err := exec.Command("getconf", s).Output()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", out), nil
}

func PurgeDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}

// WalkDir: walks the directory up to a maximum depth. When it encounters
// the specified pattern (`re`), it will append that path to the list `toPurge`.
// If any error is encountered while accessing a directory during the "walk",
// that directory is skipped.
func WalkDir(dir string, re *regexp.Regexp, toPurge *[]string, maxDepth int) {
	filesystem := os.DirFS(dir)
	fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if len((*re).FindAll([]byte(path), -1)) != 0 {
			if strings.Count(path, string(os.PathSeparator)) <= maxDepth {
				*toPurge = append(*toPurge, filepath.Join(dir, path))
			}
		}
		return nil
	})
}
