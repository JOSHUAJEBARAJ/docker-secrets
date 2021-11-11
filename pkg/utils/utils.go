package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
func Untar(tarball, target string) error {

	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func walk(s string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		fileExtension := filepath.Ext(s)

		if fileExtension == ".tar" {
			layer_file := strings.HasSuffix(s, "layer.tar")
			if layer_file {
				path := fileNameWithoutExtSliceNotation(s)
				err := Untar(s, path)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	return nil
}

// untarring output folder
func Outputar() error {
	fmt.Println("📦 Untarring the Docker Layers")
	err := filepath.WalkDir("output", walk)
	if err != nil {
		return err
	}
	return nil
}

func Scan() error {

	fmt.Println("🔍 Searching for Secrets")
	cmd := exec.Command("detect-secrets", "scan", "--all-files", "output")

	// open the out file for writing
	outfile, err := os.Create("results.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Wait()
	fmt.Println("📁 Output written in result.json file ")
	return nil
}
