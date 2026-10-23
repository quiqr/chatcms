package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

// ThumbnailWidth is how wide a picture is drawn in the library's grid. The
// designs ask for a grid of 150 pixel cells, and twice that is what a screen
// with two pixels to the point needs.
const ThumbnailWidth = 300

// Thumbnails keeps smaller copies of pictures so that a library of a hundred
// photographs does not send a hundred photographs.
//
// A thumbnail is named after the bytes it was made from, so it is a cache in
// the strict sense: throwing the whole directory away costs nothing but the
// work of making them again, and nothing in it has to be backed up or
// migrated. The site in git remains the only thing that matters.
type Thumbnails struct {
	dir string
}

func NewThumbnails(dir string) (*Thumbnails, error) {
	if dir == "" {
		return nil, fmt.Errorf("media: thumbnails need somewhere to live")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("media: %w", err)
	}
	return &Thumbnails{dir: dir}, nil
}

// Of gives a small copy of a picture, making one the first time it is asked
// for. A kind it cannot read, such as a drawing, is given back as it is:
// those are already small and scale by themselves.
func (t *Thumbnails) Of(content []byte, kind string) ([]byte, string, error) {
	if kind == "image/svg+xml" {
		return content, kind, nil
	}

	sum := sha256.Sum256(content)
	name := filepath.Join(t.dir, hex.EncodeToString(sum[:])+".png")

	if held, err := os.ReadFile(name); err == nil {
		return held, "image/png", nil
	}

	smaller, err := shrink(content)
	if err != nil {
		// A picture this cannot read is still a picture the site holds. It is
		// given back whole rather than reported as broken.
		return content, kind, nil
	}

	// Written beside and moved into place, so that two requests for the same
	// new picture cannot leave half a file behind.
	temporary, err := os.CreateTemp(t.dir, "making-")
	if err != nil {
		return smaller, "image/png", nil
	}
	if _, err := temporary.Write(smaller); err == nil {
		temporary.Close()
		os.Rename(temporary.Name(), name)
	} else {
		temporary.Close()
		os.Remove(temporary.Name())
	}
	return smaller, "image/png", nil
}

func shrink(content []byte) ([]byte, error) {
	source, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	bounds := source.Bounds()
	if bounds.Dx() <= ThumbnailWidth {
		return encode(source)
	}

	height := bounds.Dy() * ThumbnailWidth / bounds.Dx()
	if height < 1 {
		height = 1
	}
	smaller := image.NewRGBA(image.Rect(0, 0, ThumbnailWidth, height))
	xdraw.CatmullRom.Scale(smaller, smaller.Bounds(), source, bounds, draw.Over, nil)
	return encode(smaller)
}

func encode(held image.Image) ([]byte, error) {
	out := bytes.Buffer{}
	if err := png.Encode(&out, held); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// The decoders register themselves, which is what lets image.Decode read a
// picture without being told what it is.
var _ = jpeg.Decode
var _ = gif.Decode
