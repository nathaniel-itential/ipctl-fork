// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/itential/ipctl/internal/logging"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// PathExists will check if the specified fie path exists
func PathExists(fp string) bool {
	_, err := os.Stat(fp)
	return !os.IsNotExist(err)
}

// LoadObject will take in a mapped interface object and load it into the a
// instance of a struct specified by ptr
func LoadObject(in any, ptr any) {
	b, err := json.Marshal(in)
	if err != nil {
		logging.Fatal(err, "failed to marshal data")
	}

	if err := json.Unmarshal(b, ptr); err != nil {
		logging.Fatal(err, "failed to unmarshal data")
	}
}

// NormalizeFilename will take both a filename and filepath argument and
// commbine them.  It will also check the filename and replace all instances of
// "/" with "_".
func NormalizeFilename(fn, fp string) (string, error) {
	logging.Trace()

	fn = strings.Replace(fn, "/", "_", -1)

	if fp != "" {
		fn = filepath.Join(fp, fn)
	}

	dir, err := homedir.Expand(fn)
	if err != nil {
		return "", err
	}

	return dir, nil
}

// WriteBytesToDisk accepts a byte array and filepath used to write to disk.
// If the destination already exists, the function will not overwrite the file
// unless the overwrite argument is set to true.
func WriteBytesToDisk(b []byte, dst string, overwrite bool) error {
	logging.Trace()

	if err := EnsurePathExists(filepath.Dir(dst)); err != nil {
		return err
	}

	if PathExists(dst) {
		if overwrite {
			if err := os.Remove(dst); err != nil {
				return err
			}
		} else {
			return errors.New("specified destintation file already exists")
		}
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(b)
	f.Sync()

	return nil
}

// WriteJsonToDisk will take any struct object, marshal it to pretty JSON and
// then write it to disk.  The fn argument defines the name of the file to
// write.  The fp argument defines the path.  If the fp argument is empty, the
// current working directory is used.
func WriteJsonToDisk(o any, fn, fp string) error {
	logging.Trace()

	dst, err := NormalizeFilename(fn, fp)
	if err != nil {
		return err
	}
	logging.Debug("Writing file `%s` to path `%s`", fn, fp)

	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		return err
	}

	return WriteBytesToDisk(b, dst, true)
}

// WRiteYamlToDisck will accept an object, marshal it to yaml encoding and
// write it to disk.
func WriteYamlToDisk(o any, fn, fp string) error {
	logging.Trace()

	dst, err := NormalizeFilename(fn, fp)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(o)
	if err != nil {
		return err
	}

	return WriteBytesToDisk(b, dst, false)
}

// Write accepts any object and will write it to disk as specified by the
// filename and filepath.
func Write(o any, fn, fp, encoding string) error {
	logging.Trace()

	if encoding == "" {
		encoding = "json"
	}

	var err error

	switch encoding {
	case "json":
		err = WriteJsonToDisk(o, fn, fp)
	case "yaml":
		err = WriteYamlToDisk(o, fn, fp)
	}

	return err
}

func ReadFromFile(path string) ([]byte, error) {
	logging.Trace()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	return os.ReadFile(path)
}

// ReadObjectFromDisk will attempt to read a JSON or YAML encoded object from
// disk into the ptr argument.  The path argument must specify the full path
// including filename to the file to read.
func ReadObjectFromDisk(path string, ptr any) error {
	logging.Trace()

	data, err := ReadFromFile(path)
	if err != nil {
		return err
	}

	return UnmarshalData(data, ptr)
}

func ReadStringFromFile(path string) (string, error) {
	logging.Trace()
	data, err := ReadFromFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EnsurePathExists will check if the path specified by p exists and will
// create it if it does not exist.
func EnsurePathExists(p string) error {
	if !PathExists(p) {
		return os.MkdirAll(p, os.ModePerm)
	}
	return nil
}

// IsDir will check the path argument to see if it is a directory or not.
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err // Return false if the path doesn't exist or there's another error
	}
	return info.IsDir(), nil
}
