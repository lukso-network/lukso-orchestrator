package fileutil

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ExpandPath given a string which may be a relative path.
// 1. replace tilde with users home dir
// 2. expands embedded environment variables
// 3. cleans the path, e.g. /a/b/../c -> /a/c
// Note, it has limitations, e.g. ~someuser/tmp will not be expanded
func ExpandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := HomeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return filepath.Abs(path.Clean(os.ExpandEnv(p)))
}

// HomeDir for a user.
func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// HasDir checks if a directory indeed exists at the specified path.
func HasDir(dirPath string) (bool, error) {
	fullPath, err := ExpandPath(dirPath)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if info == nil {
		return false, err
	}
	return info.IsDir(), err
}

// FileExists returns true if a file is not a directory and exists
// at the specified path.
func FileExists(filename string) bool {
	filePath, err := ExpandPath(filename)
	if err != nil {
		return false
	}
	info, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.WithError(err).Info("Checking for file existence returned an error")
		}
		return false
	}
	return info != nil && !info.IsDir()
}

// RecursiveFileFind returns true, and the path,  if a file is not a directory and exists
// at  dir or any of its subdirectories.  Finds the first instant based on the Walk order and returns.
// Define non-fatal error to stop the recursive directory walk
var stopWalk = errors.New("stop walking")

func RecursiveFileFind(filename, dir string) (bool, string, error) {
	var found bool
	var fpath string
	dir = filepath.Clean(dir)
	found = false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// checks if its a file  and has the exact name as the filename
		// need to break the walk function by using a non-fatal error
		if !info.IsDir() && filename == info.Name() {
			found = true
			fpath = path
			return stopWalk
		}

		// no errors or file found
		return nil
	})
	if err != nil && err != stopWalk {
		return false, "", err
	}
	return found, fpath, nil
}

// ReadFileAsBytes expands a file name's absolute path and reads it as bytes from disk.
func ReadFileAsBytes(filename string) ([]byte, error) {
	filePath, err := ExpandPath(filename)
	if err != nil {
		return nil, errors.Wrap(err, "could not determine absolute path of password file")
	}
	return ioutil.ReadFile(filePath)
}

// DirsEqual checks whether two directories have the same content.
func DirsEqual(src, dst string) bool {
	hash1, err := HashDir(src)
	if err != nil {
		return false
	}

	hash2, err := HashDir(dst)
	if err != nil {
		return false
	}

	return hash1 == hash2
}

// HashDir calculates and returns hash of directory: each file's hash is calculated and saved along
// with the file name into the list, after which list is hashed to produce the final signature.
// Implementation is based on https://github.com/golang/mod/blob/release-branch.go1.15/sumdb/dirhash/hash.go
func HashDir(dir string) (string, error) {
	files, err := DirFiles(dir)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	files = append([]string(nil), files...)
	sort.Strings(files)
	for _, file := range files {
		fd, err := os.Open(filepath.Join(dir, file))
		if err != nil {
			return "", err
		}
		hf := sha256.New()
		_, err = io.Copy(hf, fd)
		if err != nil {
			return "", err
		}
		if err := fd.Close(); err != nil {
			return "", err
		}
		if _, err := fmt.Fprintf(h, "%x  %s\n", hf.Sum(nil), file); err != nil {
			return "", err
		}
	}
	return "hashdir:" + base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// DirFiles returns list of files found within a given directory and its sub-directories.
// Directory prefix will not be included as a part of returned file string i.e. for a file located
// in "dir/foo/bar" only "foo/bar" part will be returned.
func DirFiles(dir string) ([]string, error) {
	var files []string
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relFile := file
		if dir != "." {
			relFile = file[len(dir)+1:]
		}
		files = append(files, filepath.ToSlash(relFile))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ipcEndpoint resolves an IPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func IpcEndpoint(ipcPath, datadir string) string {
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(ipcPath, `\\.\pipe\`) {
			return ipcPath
		}
		return `\\.\pipe\` + ipcPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(ipcPath) == ipcPath {
		if datadir == "" {
			return filepath.Join(os.TempDir(), ipcPath)
		}
		return filepath.Join(datadir, ipcPath)
	}
	return ipcPath
}
