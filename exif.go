package ma

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

var _ Exif = (*GoExif)(nil)

const exifHeaderSize = 4

// MetaData represents the EXIF data about a file
type MetaData struct {
	// Info is the analyzed file
	Info fs.FileInfo

	// Err is non-nil if an error occurred processing the file
	Err error
	// DateTime is the best effort `DateTimeOriginal` of the file
	DateTime time.Time
}

// Exif extracts EXIF metadata from files
type Exif interface {
	// Extract returns metadata about a file
	Extract(afs afero.Fs, dirname string, infos ...fs.FileInfo) []MetaData
}

type GoExif struct{}

func (x *GoExif) datetime(afs afero.Fs, filename string) (time.Time, error) {
	fp, err := afs.Open(filename)
	if err != nil {
		return time.Time{}, err
	}
	defer fp.Close()
	m, err := exif.Decode(fp)
	if err != nil {
		return time.Time{}, err
	}
	return m.DateTime()
}

func (x *GoExif) Extract(afs afero.Fs, dirname string, infos ...fs.FileInfo) []MetaData {
	mds := make([]MetaData, len(infos))
	for i := range infos {
		mds[i] = MetaData{Info: infos[i]}
		ext := strings.ToLower(filepath.Ext(mds[i].Info.Name()))
		switch ext {
		case "", ".pxm", ".xmp":
		case ".orf", ".mov", ".avi", ".mp4":
			mds[i].DateTime = mds[i].Info.ModTime()
		default:
			if mds[i].Info.Size() < exifHeaderSize {
				// the exif header is four bytes long so bail rather than EOF
				continue
			}
			mds[i].DateTime, mds[i].Err = x.datetime(afs, filepath.Join(dirname, mds[i].Info.Name()))
		}
	}
	return mds
}

func xif(c *cli.Context) error {
	afs := runtime(c).Fs
	dtr := runtime(c).Exif
	for i := 0; i < c.NArg(); i++ {
		if err := afero.Walk(afs, c.Args().Get(i), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			dirname, _ := filepath.Split(path)
			for _, m := range dtr.Extract(afs, dirname, info) {
				if m.Err != nil {
					if errors.Is(m.Err, io.EOF) {
						log.Warn().Time("datetime", m.DateTime).Str("filename", path).Msg(c.Command.Name)
						return nil
					}
					log.Err(m.Err).Time("datetime", m.DateTime).Str("filename", path).Msg(c.Command.Name)
					return m.Err
				}
				log.Info().Str("filename", path).Time("datetime", m.DateTime).Msg(c.Command.Name)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func CommandExif() *cli.Command {
	return &cli.Command{
		Name:        "exif",
		HelpName:    "exif",
		Hidden:      true,
		Usage:       "debugging tool for exif data",
		Description: "debugging tool for exif data",
		Action:      xif,
	}
}
