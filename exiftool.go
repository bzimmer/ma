//go:build exiftool

package ma

/*
This implementation of an `Exif` leverages the external `exiftool`. While `exiftool`
is definitely more capable of a wide range of files it's also far slower and requires an
external dependency.
*/

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/spf13/afero"
)

var _ Exif = (*perl)(nil)

const exifTimeLayout = "2006:01:02 15:04:05"

type ExiftoolOption func(*perl) error

func NewExiftool(options ...ExiftoolOption) (Exif, error) {
	x := new(perl)
	for _, opt := range options {
		if err := opt(x); err != nil {
			return nil, err
		}
	}
	if x.tool == nil {
		return nil, errors.New("no exiftool.Tool specified")
	}
	return x, nil
}

type perl struct {
	tool *exiftool.Exiftool
}

func (x *perl) Extract(_ afero.Fs, dirname string, infos ...fs.FileInfo) []MetaData {
	filenames := make([]string, len(infos))
	for i := range infos {
		switch ext := strings.ToLower(filepath.Ext(infos[i].Name())); ext {
		case "", ".pxm", ".xmp":
		default:
			filenames[i] = filepath.Join(dirname, infos[i].Name())
		}
	}
	mds := make([]MetaData, len(infos))

	for i, m := range x.tool.ExtractMetadata(filenames...) {
		if m.Err != nil {
			if filenames[i] != "" {
				mds[i].Err = m.Err
			}
			continue
		}
		if dto, ok := m.Fields["DateTimeOriginal"]; ok {
			timeZone := time.Local
			tm, err := time.ParseInLocation(exifTimeLayout, dto.(string), timeZone)
			if err != nil {
				mds[i].Err = err
				continue
			}
			mds[i].DateTime = tm
		}
	}
	return mds
}
