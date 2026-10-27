package auth_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoPictureOfAPersonCanReachASite is the rule the storage decision
// rests on: a person's picture belongs to the account, crosses
// organisations, and a customer's published site must never carry one.
//
// It is enforced by where the bytes live rather than by a check at publish
// time, so this asserts the shape: nothing that writes into a site ever
// reads a person's picture.
func TestNoPictureOfAPersonCanReachASite(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}

	// The packages that write into a site, and so could carry one in.
	for _, where := range []string{
		"internal/gitsync",
		"internal/publish",
		"internal/content",
		"internal/media",
		"internal/sitebuild",
	} {
		err := filepath.WalkDir(filepath.Join(root, where), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return walkErr
			}

			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, forbidden := range []string{"ReadPicture", "SetPicture", "picture_kind"} {
				if strings.Contains(string(source), forbidden) {
					t.Errorf("%s names %s; a person's picture must never reach a site",
						strings.TrimPrefix(path, root+"/"), forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s failed: %v", where, err)
		}
	}
}
