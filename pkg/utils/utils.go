package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
func create_gz(fn string) error {
	f, err := os.Create(fn + ".gz")

	if err != nil {
		return err
	}

	in, err := ioutil.ReadFile(fn)

	if err != nil {
		return err
	}

	w := gzip.NewWriter(f)

	w.Write(in)

	w.Close()

	return nil
}

func Unload(src, dest string) error {
	// creating the .gz file
	create_err := create_gz(src)
	if create_err != nil {
		return create_err
	}
	file, err := os.Open(src + ".gz")
	if err != nil {
		return err
	}
	untar_err := Untar(file, dest)
	if err != nil {
		return untar_err
	}
	// deleting file after decompression
	defer delete_gz(src)
	return nil
}

func delete_gz(src string) {
	e := os.Remove(src + ".gz")
	if e != nil {
		log.Fatal(e)
	}
}

// code taken from https://github.com/k3s-io/k3s/blob/v1.0.1/pkg/untar/untar.go
func Untar(r io.Reader, dir string) error {
	return untar(r, dir)
}

func untar(r io.Reader, dir string) (err error) {
	t0 := time.Now()
	nFiles := 0
	madeDir := map[string]bool{}
	defer func() {
		td := time.Since(t0)
		if err != nil {
			logrus.Printf("error extracting tarball into %s after %d files, %d dirs, %v: %v", dir, nFiles, len(madeDir), td, err)
		}
	}()
	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("requires gzip-compressed body: %v", err)
	}
	tr := tar.NewReader(zr)
	loggedChtimesError := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Printf("tar reading error: %v", err)
			return fmt.Errorf("tar error: %v", err)
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		rel := filepath.FromSlash(f.Name)
		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
					return err
				}
				madeDir[dir] = true
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			modTime := f.ModTime
			if modTime.After(t0) {
				// Clamp modtimes at system time. See
				// golang.org/issue/19062 when clock on
				// buildlet was behind the gitmirror server
				// doing the git-archive.
				modTime = t0
			}
			if !modTime.IsZero() {
				if err := os.Chtimes(abs, modTime, modTime); err != nil && !loggedChtimesError {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					logrus.Printf("error changing modtime: %v (further Chtimes errors suppressed)", err)
					loggedChtimesError = true // once is enough
				}
			}
			nFiles++
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
			madeDir[abs] = true
		case f.Linkname != "":
			if err := os.Symlink(f.Linkname, abs); err != nil {

				// to do
				continue
			}
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %v", f.Name, mode)
		}
	}
	return nil
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
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
				fmt.Println(s)
				path := fileNameWithoutExtSliceNotation(s)
				err := Unload(s, path)
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
	fmt.Println("📁 Output written in results.json file ")
	return nil
}
