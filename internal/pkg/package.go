package pkg

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/go-pkg-org/gopkg/internal/util"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileExt is the extension for package files
const (
	FileExt   = "pkg"
	srcSuffix = "src"
)

// Type represent a package type
type Type string

const (
	// Control are package used to build source & binary
	Control Type = "control"
	// Source are package providing Go source code
	Source Type = "source"
	// Binary are package providing executable
	Binary Type = "binary"
)

// Entry is a tiny struct to contain data for a specific
// entry that will be archived into a pkg file.
type Entry struct {
	FilePath    string
	ArchivePath string
}

// CreateEntries creates a slice with all files in a specific directory that should be added to the archive.
// The resulting value is a Entry, which maps a filepath to an archive path.
func CreateEntries(path string, pathPrefix string, excludedFiles []string) ([]Entry, error) {
	dirContent, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Write file list.
	var fileList []Entry
	for _, file := range dirContent {
		if util.Contains(excludedFiles, file.Name()) {
			log.Trace().Str("file", filepath.Join(path, file.Name())).Msg("Skipping file")
			continue
		}

		if file.IsDir() {
			tmp, err := CreateEntries(filepath.Join(path, file.Name()), "", excludedFiles)
			if err != nil {
				return nil, err
			}
			for _, p := range tmp {
				fileList = append(fileList, Entry{
					FilePath:    p.FilePath,
					ArchivePath: filepath.Join(pathPrefix, file.Name(), p.ArchivePath),
				})
			}
		} else {
			fileList = append(fileList, Entry{
				FilePath:    filepath.Join(path, file.Name()),
				ArchivePath: filepath.Join(pathPrefix, file.Name()),
			})
		}
	}

	return fileList, nil
}

// Read reads a package and returns content.
func Read(path string) (map[string][]byte, error) {
	result := map[string][]byte{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(file)

	tr := tar.NewReader(buffer)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		out := bytes.NewBuffer([]byte{})
		if _, err := out.ReadFrom(tr); err != nil {
			return nil, err
		}

		result[header.Name] = out.Bytes()
	}
	return result, nil
}

// Write creates a tar file from a set of ArchiveEntries.
func Write(path string, files []Entry, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("failed to create new tar source (file already exist)")
		}
	}

	var buffer bytes.Buffer
	tw := tar.NewWriter(&buffer)

	for _, file := range files {
		log.Trace().Str("file-path", file.FilePath).Str("archive-path", file.ArchivePath).Msg("Writing file")
		fileBody, err := ioutil.ReadFile(file.FilePath)
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name: file.ArchivePath,
			Mode: 0644,
			Size: int64(len(fileBody)),
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write(fileBody); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return ioutil.WriteFile(path, buffer.Bytes(), 0644)
}

// GetFileName return package file name corresponding to given information
func GetFileName(name, version, os, arch string, pkgType Type) (string, error) {
	// validate common information
	if name == "" || version == "" {
		return "", fmt.Errorf("missing information to build package name")
	}

	pkgName := GetName(name, pkgType == Source)
	switch pkgType {
	case Control:
		return fmt.Sprintf("%s_%s.%s", pkgName, version, FileExt), nil
	case Source:
		return fmt.Sprintf("%s_%s.%s", pkgName, version, FileExt), nil
	case Binary:
		// validate binary specific information
		if os == "" || arch == "" {
			return "", fmt.Errorf("missing information to build package name")
		}

		return fmt.Sprintf("%s_%s_%s_%s.%s", pkgName, version, os, arch, FileExt), nil
	default:
		return "", fmt.Errorf("non managed packaged type: %s", pkgType)
	}
}

// ParseFileName extract package information from file name
func ParseFileName(fileName string) (string, string, string, string, Type, error) {
	fileName = strings.TrimSuffix(fileName, "."+FileExt)

	parts := strings.Split(fileName, "_")
	switch len(parts) {
	case 2:
		if strings.HasSuffix(parts[0], "-"+srcSuffix) {
			return parts[0], parts[1], "", "", Source, nil
		}
		return parts[0], parts[1], "", "", Control, nil
	case 4:
		return parts[0], parts[1], parts[2], parts[3], Binary, nil
	default:
		return "", "", "", "", "", fmt.Errorf("invalid package: %s", fileName)
	}
}

// GetName translate from importPath to package name
func GetName(importPath string, isSrc bool) string {
	name := strings.ReplaceAll(importPath, "/", "-")
	if isSrc {
		name = fmt.Sprintf("%s-%s", name, srcSuffix)
	}

	return name
}
