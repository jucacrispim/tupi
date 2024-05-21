// Copyright 2020 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi.

// tupi is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package tupi

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var chunkSize int64 = 10 << 20

func genRandFname(fname string) (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-", b) + fname, nil
}

// writeFile writes the contents of an uploaded file into a file in the
// local fs.
func writeFile(dir string, r *multipart.Reader, randfname bool, prevent_overwrite bool) (string, error) {

	part, err := r.NextPart()
	if err != nil {
		return "", err
	}
	var fname string
	if randfname {
		fname, err = genRandFname(part.FileName())
		if err != nil {
			return "", err
		}
	} else {
		fname = part.FileName()
	}

	fpath := dir + string(os.PathSeparator) + fname
	if fileExists(fpath) && prevent_overwrite {
		return "", errors.New("File " + fname + " already exists")
	}

	AcquireLock(fpath)
	defer ReleaseLock(fpath)
	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var off int64 = 0
	for {
		c, err := ioutil.ReadAll(part)
		if err != nil {
			return fname, err
		}
		f.WriteAt([]byte(c), off)
		err = f.Sync()
		if err != nil {
			return fname, err
		}
		off += int64(len(c))

		part, err = r.NextPart()

		if err == io.EOF {
			break
		}

	}
	return fname, nil
}

func fileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	if err == nil {
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}

// extractFiles extract the contents of a tar.gz file to the local
// file system. All files will be extracted inside `root_dir`
func extractFiles(file io.Reader, root_dir string, prevent_overwrite bool) ([]string, error) {
	buf, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer buf.Close()
	tr := tar.NewReader(buf)
	files := make([]string, 0)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		fname := hdr.Name
		path := filepath.Join(root_dir, fname)
		switch hdr.Typeflag {
		case tar.TypeDir:
			// for a directory we hold the lock till the end of the function
			// to avoid someome messing with a directory we are working inside
			AcquireLock(path)
			defer ReleaseLock(path)
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return nil, err
			}
			files = append(files, fname)

		case tar.TypeReg:
			if fileExists(path) && prevent_overwrite {
				return nil, errors.New("File " + path + " already exists")
			}
			AcquireLock(path)
			out, err := os.Create(path)
			if err != nil {
				ReleaseLock(path)
				return nil, err
			}
			_, err = io.Copy(out, tr)
			ReleaseLock(path)
			if err != nil {
				return nil, err
			}
			files = append(files, fname)

		case tar.TypeSymlink:
			target := filepath.Join(filepath.Dir(path), hdr.Linkname)
			// if the symlink points to a file outside of the root_dir
			// we append the root_dir to it, basically breaking the link
			if !strings.HasPrefix(target, root_dir) {
				target = filepath.Join(root_dir, strings.TrimLeft(target, "/"))
			}

			AcquireLock(path)
			err := os.Symlink(target, path)
			ReleaseLock(path)
			if err != nil {
				return nil, err
			}

			files = append(files, fname)

		default:
			// notest
			log.Printf("Unknown type %d for %s", hdr.Typeflag, path)
		}

	}
	return files, nil
}

// from now on it a copy with modifications from the http package code
// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) {
	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		// notest
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	url := r.URL.Path
	// redirect if the directory name doesn't end in a slash
	if d.IsDir() && (url == "" || url[len(url)-1] != '/') {
		http.Redirect(w, r, path.Base(url)+"/", http.StatusMovedPermanently)
		return
	}

	// List the contents of a directory
	if d.IsDir() {
		if checkIfModifiedSince(r, d.ModTime()) == condFalse {
			// if content hasn't been modified
			// returns 304 not modified
			writeNotModified(w)
			return
		}
		setLastModified(w, d.ModTime())
		dirList(w, r, f)
		return
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value. It's important that toHTTPError does not
// actually return err.Error(), since msg and httpStatus are returned to users,
// and historically Go's ServeContent always returned just "404 Not Found" for
// all errors. We don't want to start leaking information in error messages.
func toHTTPError(err error) (msg string, httpStatus int) {
	if errors.Is(err, fs.ErrNotExist) {
		return "404 page not found", http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError // notest
}

// condResult is the result of an HTTP request precondition check.
// See https://tools.ietf.org/html/rfc7232 section 3.
type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

func checkIfModifiedSince(r *http.Request, modtime time.Time) condResult {
	ims := r.Header.Get("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return condNone
	}
	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return condFalse
	}
	return condTrue
}

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(time.Unix(0, 0))
}

func writeNotModified(w http.ResponseWriter) {
	h := w.Header()
	delete(h, "Content-Type")
	delete(h, "Content-Length")
	delete(h, "Content-Encoding")
	w.WriteHeader(http.StatusNotModified)
}

func setLastModified(w http.ResponseWriter, modtime time.Time) {
	if !isZeroTime(modtime) {
		w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	}
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

type anyDirs interface {
	len() int
	name(i int) string
	isDir(i int) bool
}

type fileInfoDirs []fs.FileInfo

func (d fileInfoDirs) len() int          { return len(d) }       // notest
func (d fileInfoDirs) isDir(i int) bool  { return d[i].IsDir() } // notest
func (d fileInfoDirs) name(i int) string { return d[i].Name() }  // notest

type dirEntryDirs []fs.DirEntry

func (d dirEntryDirs) len() int          { return len(d) }
func (d dirEntryDirs) isDir(i int) bool  { return d[i].IsDir() }
func (d dirEntryDirs) name(i int) string { return d[i].Name() }

func dirList(w http.ResponseWriter, r *http.Request, f http.File) {
	// Prefer to use ReadDir instead of Readdir,
	// because the former doesn't require calling
	// Stat on every entry of a directory on Unix.
	var dirs anyDirs
	var err error
	if d, ok := f.(fs.ReadDirFile); ok {
		var list dirEntryDirs
		list, err = d.ReadDir(-1)
		dirs = list
	} else {
		// notest
		var list fileInfoDirs
		list, err = f.Readdir(-1)
		dirs = list
	}

	if err != nil {
		Errorf("http: error reading directory: %v", err)
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs.name(i) < dirs.name(j) })

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for i, n := 0, dirs.len(); i < n; i++ {
		name := dirs.name(i)
		if dirs.isDir(i) {
			name += "/"
		}
		// name may contain '?' or '#', which must be escaped to remain
		// part of the URL path, and not indicate the start of a query
		// string or fragment.
		url := url.URL{Path: name}
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
	}
	fmt.Fprintf(w, "</pre>\n")
}
