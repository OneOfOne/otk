package otk

import (
	"encoding/base64"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

var ErrInvalidImageDataURL = xerrors.New("invalid dataURL, must be in format `data:image/{png,jpeg,jpg,gif,webp};base64,`")

func ImageFromDataURL(dataURL string, fn func(ext string) (io.WriteCloser, error)) error {
	const (
		imgHeaderBase   = "data:image/"
		imgHeaderSuffix = ";base64,"
	)

	if !strings.HasPrefix(dataURL, imgHeaderBase) || len(dataURL) < len(imgHeaderBase)+5 {
		return ErrInvalidImageDataURL
	}

	ext := dataURL[len(imgHeaderBase) : len(imgHeaderBase)+4]
	switch ext {
	case "png;", "gif;", "webp;":
		ext = ext[:3]
	case "jpg;", "jpeg":
		ext = "jpg"
	default:
		return ErrInvalidImageDataURL
	}

	i := strings.Index(dataURL, imgHeaderSuffix)
	if i == -1 {
		return ErrInvalidImageDataURL
	}

	wc, err := fn(ext)
	if err != nil {
		return err
	}
	defer wc.Close()

	rd := strings.NewReader(dataURL[i+len(imgHeaderSuffix):])
	_, err = io.Copy(wc, base64.NewDecoder(base64.StdEncoding, rd))
	return err
}

// SaveImageFromDataURL expects output name without the extension
// returns full path
func SaveImageFromDataURL(dataURL, output string) (fp string, err error) {
	os.MkdirAll(filepath.Dir(output), 0755)
	err = ImageFromDataURL(dataURL, func(ext string) (io.WriteCloser, error) {
		fp = output + "." + ext
		return os.Create(fp)
	})
	return
}

// DecodeImageConfig accepts a reader and returns a CachedReader and the results of image.DecodeConfig
func DecodeImageConfig(rd io.Reader) (safeReader io.Reader, cfg image.Config, format string, err error) {
	cr := &CachedReader{R: rd}
	cfg, format, err = image.DecodeConfig(cr)
	cr.Rewind()
	safeReader = cr
	return
}
