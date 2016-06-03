// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package bindata

import (
	"fmt"
	"io"
)

func writeRestore(w io.Writer) error {
	_, err := fmt.Fprintf(w, `
// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	return save(_filePath(dir, name), data, info.Mode(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

// save saves the given data to the file at filename. If an existing file at
// that filename already exists, this simply chmods the existing file to match
// the given fileMode and otherwise leaves it alone.
func save(filename string, data []byte, fileMode os.FileMode, modTime time.Time) error {
	path := filepath.Dir(filename)
	if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}

		if dataMatches(filename, data) {
			err2 := chmod(filename, fileMode)
			if err2 != nil {
				return err2
			}
			return os.Chtimes(filename, modTime, modTime)
		}

		file, err = openAndTruncate(filename, fileMode, true)
		if err != nil {
			return err
		}
	}

	if _, err = file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return os.Chtimes(filename, modTime, modTime)
}

func openAndTruncate(filename string, fileMode os.FileMode, removeIfNecessary bool) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil && os.IsPermission(err) && removeIfNecessary {
		if err = os.Remove(filename); err != nil {
			return nil, err
		}
		return openAndTruncate(filename, fileMode, false)
	}

	return file, err
}

// dataMatches compares the file at filename byte for byte with the given data
func dataMatches(filename string, data []byte) bool {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}
	if fileInfo.Size() != int64(len(data)) {
		return false
	}
	b := make([]byte, 65536)
	i := 0
	for {
		n, err := file.Read(b)
		if err != nil && err != io.EOF {
			return false
		}
		for j := 0; j < n; j++ {
			if b[j] != data[i] {
				return false
			}
			i = i + 1
		}
		if err == io.EOF {
			break
		}
	}
	return true
}

func chmod(filename string, fileMode os.FileMode) error {
	fi, err := os.Stat(filename)
	if err != nil || fi.Mode() != fileMode {
		return os.Chmod(filename, fileMode)
	}
	return err
}

`)
	return err
}
