package fileutil_test

import (
	"bufio"
	"bytes"
	"github.com/lukso-network/lukso-orchestrator/shared/fileutil"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/require"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"testing"
)

func TestPathExpansion(t *testing.T) {
	user, err := user.Current()
	require.NoError(t, err)
	tests := map[string]string{
		"/home/someuser/tmp": "/home/someuser/tmp",
		"~/tmp":              user.HomeDir + "/tmp",
		"$DDDXXX/a/b":        "/tmp/a/b",
		"/a/b/":              "/a/b",
	}
	require.NoError(t, os.Setenv("DDDXXX", "/tmp"))
	for test, expected := range tests {
		expanded, err := fileutil.ExpandPath(test)
		require.NoError(t, err)
		assert.Equal(t, expected, expanded)
	}
}

func TestHashDir(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		hash, err := fileutil.HashDir(filepath.Join(t.TempDir(), "nonexistent"))
		assert.ErrorContains(t, "no such file or directory", err)
		assert.Equal(t, "", hash)
	})

	t.Run("empty directory", func(t *testing.T) {
		hash, err := fileutil.HashDir(t.TempDir())
		assert.NoError(t, err)
		assert.Equal(t, "hashdir:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=", hash)
	})

	t.Run("non-empty directory", func(t *testing.T) {
		tmpDir, _ := tmpDirWithContents(t)
		hash, err := fileutil.HashDir(tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, "hashdir:oSp9wRacwTIrnbgJWcwTvihHfv4B2zRbLYa0GZ7DDk0=", hash)
	})
}

func TestDirFiles(t *testing.T) {
	tmpDir, tmpDirFnames := tmpDirWithContents(t)
	tests := []struct {
		name     string
		path     string
		outFiles []string
	}{
		{
			name:     "dot path",
			path:     filepath.Join(tmpDir, "/./"),
			outFiles: tmpDirFnames,
		},
		{
			name:     "non-empty folder",
			path:     tmpDir,
			outFiles: tmpDirFnames,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outFiles, err := fileutil.DirFiles(tt.path)
			require.NoError(t, err)

			sort.Strings(outFiles)
			assert.DeepEqual(t, tt.outFiles, outFiles)
		})
	}
}

func TestRecursiveFileFind(t *testing.T) {
	tmpDir, _ := tmpDirWithContentsForRecursiveFind(t)
	tests := []struct {
		name  string
		root  string
		path  string
		found bool
	}{
		{
			name:  "file1",
			root:  tmpDir,
			path:  "subfolder1/subfolder11/file1",
			found: true,
		},
		{
			name:  "file2",
			root:  tmpDir,
			path:  "subfolder2/file2",
			found: true,
		},
		{
			name:  "file1",
			root:  tmpDir + "/subfolder1",
			path:  "subfolder11/file1",
			found: true,
		},
		{
			name:  "file3",
			root:  tmpDir,
			path:  "file3",
			found: true,
		},
		{
			name:  "file4",
			root:  tmpDir,
			path:  "",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, _, err := fileutil.RecursiveFileFind(tt.name, tt.root)
			require.NoError(t, err)

			assert.DeepEqual(t, tt.found, found)
		})
	}
}

func deepCompare(t *testing.T, file1, file2 string) bool {
	sf, err := os.Open(file1)
	assert.NoError(t, err)
	df, err := os.Open(file2)
	assert.NoError(t, err)
	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	for sscan.Scan() && dscan.Scan() {
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			return false
		}
	}
	return true
}

// tmpDirWithContents returns path to temporary directory having some folders/files in it.
// Directory is automatically removed by internal testing cleanup methods.
func tmpDirWithContents(t *testing.T) (string, []string) {
	dir := t.TempDir()
	fnames := []string{
		"file1",
		"file2",
		"subfolder1/file1",
		"subfolder1/file2",
		"subfolder1/subfolder11/file1",
		"subfolder1/subfolder11/file2",
		"subfolder1/subfolder12/file1",
		"subfolder1/subfolder12/file2",
		"subfolder2/file1",
	}
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subfolder1", "subfolder11"), 0777))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subfolder1", "subfolder12"), 0777))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subfolder2"), 0777))
	for _, fname := range fnames {
		require.NoError(t, ioutil.WriteFile(filepath.Join(dir, fname), []byte(fname), 0777))
	}
	sort.Strings(fnames)
	return dir, fnames
}

// tmpDirWithContentsForRecursiveFind returns path to temporary directory having some folders/files in it.
// Directory is automatically removed by internal testing cleanup methods.
func tmpDirWithContentsForRecursiveFind(t *testing.T) (string, []string) {
	dir := t.TempDir()
	fnames := []string{
		"subfolder1/subfolder11/file1",
		"subfolder2/file2",
		"file3",
	}
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subfolder1", "subfolder11"), 0777))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subfolder2"), 0777))
	for _, fname := range fnames {
		require.NoError(t, ioutil.WriteFile(filepath.Join(dir, fname), []byte(fname), 0777))
	}
	sort.Strings(fnames)
	return dir, fnames
}
