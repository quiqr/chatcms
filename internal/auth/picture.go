package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"

	"github.com/google/uuid"
	xdraw "golang.org/x/image/draw"

	"github.com/quiqr/chatcms/internal/imagekind"
	"github.com/quiqr/chatcms/internal/store"
)

// PictureSize is how wide and tall a person's picture is kept. It is shown
// at 32 px and 44 px and nowhere larger, so keeping the original would mean
// storing megabytes to draw something small.
const PictureSize = 128

// LargestPicture is the most that will be read. Somebody's phone produces
// several megabytes; this is refused before it is read rather than after.
const LargestPicture = 5 << 20

var (
	// ErrNotAPicture says the bytes are not a picture, whatever the file was
	// called.
	ErrNotAPicture = errors.New("that is not a picture")

	// ErrPictureTooLarge says it is bigger than this installation takes.
	ErrPictureTooLarge = fmt.Errorf("a picture may be at most %d megabytes", LargestPicture>>20)
)

// Picture is somebody's picture and what kind it is.
type Picture struct {
	Content []byte
	Kind    string
}

// SetPicture gives somebody a picture, made square and small on the way in.
func (p *People) SetPicture(ctx context.Context, personID uuid.UUID, content []byte) error {
	if len(content) > LargestPicture {
		return ErrPictureTooLarge
	}

	// What a file claims to be is not evidence; the bytes decide. This is the
	// same rule the media library follows, and the same function.
	kind, _, err := imagekind.Of(content)
	if err != nil {
		return ErrNotAPicture
	}
	if kind == "image/svg+xml" {
		// A drawing can carry script, and this one is shown inside the
		// application rather than on a published site.
		return ErrNotAPicture
	}

	square, err := squareOf(content)
	if err != nil {
		return ErrNotAPicture
	}

	return p.db.Queries.SetPicture(ctx, store.SetPictureParams{
		ID:          pgUUID(personID),
		Picture:     square,
		PictureKind: "image/png",
	})
}

// RemovePicture puts somebody back to their initials.
func (p *People) RemovePicture(ctx context.Context, personID uuid.UUID) error {
	return p.db.Queries.SetPicture(ctx, store.SetPictureParams{ID: pgUUID(personID)})
}

// ReadPicture is somebody's picture, or nothing when they have none.
func (p *People) ReadPicture(ctx context.Context, personID uuid.UUID) (Picture, bool, error) {
	row, err := p.db.Queries.ReadPicture(ctx, pgUUID(personID))
	if err != nil {
		return Picture{}, false, fmt.Errorf("reading the picture failed: %w", err)
	}
	if len(row.Picture) == 0 {
		return Picture{}, false, nil
	}
	return Picture{Content: row.Picture, Kind: row.PictureKind}, true, nil
}

// squareOf cuts the middle out and scales it, so that a wide photograph does
// not become a squashed one.
func squareOf(content []byte) ([]byte, error) {
	source, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}

	bounds := source.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	middle := image.Rect(
		bounds.Min.X+(bounds.Dx()-side)/2,
		bounds.Min.Y+(bounds.Dy()-side)/2,
		bounds.Min.X+(bounds.Dx()-side)/2+side,
		bounds.Min.Y+(bounds.Dy()-side)/2+side,
	)

	square := image.NewRGBA(image.Rect(0, 0, PictureSize, PictureSize))
	xdraw.CatmullRom.Scale(square, square.Bounds(), source, middle, xdraw.Over, nil)

	out := bytes.Buffer{}
	if err := png.Encode(&out, square); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
