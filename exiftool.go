//go:build exiftool

package ma

/*
This implementation of a `DateTimer` leverages the external `exiftool`. While `exiftool`
is definitely more capable of a wide range of files it's also far slower and requires an
external dependency.
*/

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/spf13/afero"
)

var _ DateTimer = (*Exiftool)(nil)

const exifTimeLayout = "2006:01:02 15:04:05"

type Exiftool struct {
	Tool *exiftool.Exiftool
}

func (x *Exiftool) Name() string { return "exiftool" }

func (x *Exiftool) DateTime(_ afero.Fs, dirname string, infos ...fs.FileInfo) []MetaData {
	filenames := make([]string, len(infos))
	for i := range infos {
		switch ext := strings.ToLower(filepath.Ext(infos[i].Name())); ext {
		case "", ".pxm", ".xmp":
		default:
			filenames[i] = filepath.Join(dirname, infos[i].Name())
		}
	}
	mds := make([]MetaData, len(infos))

	for i, m := range x.Tool.ExtractMetadata(filenames...) {
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
