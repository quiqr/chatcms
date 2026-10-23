package media_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms/internal/gitsync"
	"github.com/quiqr/chatcms/internal/httpapi"
	"github.com/quiqr/chatcms/internal/media"
	"github.com/quiqr/chatcms/internal/model"
	"github.com/quiqr/chatcms/internal/policy"
)

const organisation = "11111111-1111-1111-1111-111111111111"

// aPNG makes a real picture of the given size, so that what is tested is a
// picture rather than something shaped like one.
func aPNG(t *testing.T, width, height int) []byte {
	t.Helper()

	held := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			held.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 120, A: 255})
		}
	}

	out := bytes.Buffer{}
	if err := png.Encode(&out, held); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func aModel() *model.Model {
	return &model.Model{
		Singles:     []model.Single{{Key: "home", Title: "Home page", File: "content/_index.md"}},
		Collections: []model.Collection{{Key: "news", Title: "News", Folder: "content/news"}},
	}
}

type site struct {
	files   fstest.MapFS
	saved   []gitsync.Save
	changes []gitsync.Change
	model   *model.Model
}

func (s *site) FS() fs.FS { return s.files }

func (s *site) Save(ctx context.Context, save gitsync.Save) (gitsync.Version, error) {
	s.saved = append(s.saved, save)
	for name, content := range save.Writes {
		s.files[name] = &fstest.MapFile{Data: content}
	}
	for _, name := range save.Deletes {
		delete(s.files, name)
	}
	return gitsync.Version{ID: "a-version"}, nil
}

func (s *site) Changes(ctx context.Context) ([]gitsync.Change, error) { return s.changes, nil }

func (s *site) Model() (*model.Model, error) {
	if s.model == nil {
		return nil, model.ErrNoModel
	}
	return s.model, nil
}

type sources struct {
	site   *site
	closed int
}

func (s *sources) Site(ctx context.Context, projectID uuid.UUID) (media.Site, func() error, error) {
	return s.site, func() error { s.closed++; return nil }, nil
}

func (s *sources) Resource(ctx context.Context, projectID uuid.UUID) (policy.Resource, error) {
	return policy.Project(organisation, projectID.String()), nil
}

func (s *sources) Who(ctx context.Context, personID uuid.UUID) (gitsync.Identity, error) {
	return gitsync.Identity{Name: "Anna", Email: "anna@example.org"}, nil
}

type syncer struct{ asked int }

func (s *syncer) SyncSoon(ctx context.Context, projectID uuid.UUID) error {
	s.asked++
	return nil
}

func aSite(t *testing.T) *site {
	t.Helper()

	return &site{
		files: fstest.MapFS{
			"content/_index.md":       {Data: []byte("---\ntitle: Home\nhero: /images/logo.png\n---\n")},
			"content/news/one.md":     {Data: []byte("---\ntitle: One\n---\n\n![a loaf](/images/bread.png)\n")},
			"static/images/logo.png":  {Data: aPNG(t, 40, 40)},
			"static/images/bread.png": {Data: aPNG(t, 60, 30)},
			"static/images/spare.png": {Data: aPNG(t, 20, 20)},
		},
		model: aModel(),
	}
}

func anEditor(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleEditor},
	}
}

func aViewer(project uuid.UUID) policy.Actor {
	return policy.Actor{
		PersonID: uuid.NewString(),
		Projects: map[string]policy.Role{project.String(): policy.RoleViewer},
	}
}

func library(t *testing.T, src *sources, sync *syncer) *media.Library {
	t.Helper()

	thumbnails, err := media.NewThumbnails(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	made, err := media.NewLibrary(src, sync, thumbnails)
	if err != nil {
		t.Fatal(err)
	}
	return made
}

func named(pictures []media.Image, name string) (media.Image, bool) {
	for _, picture := range pictures {
		if picture.Name == name {
			return picture, true
		}
	}
	return media.Image{}, false
}

func TestEveryPictureIsListedWithWhatUsesIt(t *testing.T) {
	project := uuid.New()

	pictures, err := library(t, &sources{site: aSite(t)}, &syncer{}).
		List(context.Background(), anEditor(project), project)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(pictures) != 3 {
		t.Fatalf("listed %d pictures, want three", len(pictures))
	}

	logo, _ := named(pictures, "logo.png")
	if logo.Path != "static/images/logo.png" || logo.Size == 0 {
		t.Errorf("logo = %+v", logo)
	}
	if len(logo.UsedOn) != 1 || logo.UsedOn[0] != "home" {
		t.Errorf("logo.UsedOn = %v, want the home page", logo.UsedOn)
	}

	bread, _ := named(pictures, "bread.png")
	if len(bread.UsedOn) != 1 || bread.UsedOn[0] != "news/one" {
		t.Errorf("bread.UsedOn = %v, want the news item", bread.UsedOn)
	}

	spare, _ := named(pictures, "spare.png")
	if spare.Used() {
		t.Errorf("spare.UsedOn = %v, want nothing", spare.UsedOn)
	}
}

func TestAPictureIsAdded(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	sync := &syncer{}

	added, err := library(t, src, sync).Add(context.Background(), anEditor(project), project,
		"Our New Shop.PNG", aPNG(t, 30, 30))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if added.Name != "our-new-shop.png" {
		t.Errorf("Name = %q, want it tidied", added.Name)
	}
	if added.Path != "static/images/our-new-shop.png" {
		t.Errorf("Path = %q", added.Path)
	}
	if _, wrote := src.site.saved[0].Writes["static/images/our-new-shop.png"]; !wrote {
		t.Errorf("wrote %v", src.site.saved[0].Writes)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times, want once", sync.asked)
	}
}

func TestAPictureIsPutBesideOneAlreadyThere(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}

	added, err := library(t, src, &syncer{}).Add(context.Background(), anEditor(project), project,
		"logo.png", aPNG(t, 30, 30))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if added.Name != "logo-2.png" {
		t.Errorf("Name = %q, want it beside the one already there", added.Name)
	}
	if _, still := src.site.files["static/images/logo.png"]; !still {
		t.Error("the picture already there was replaced")
	}
}

func TestWhatIsNotAPictureIsRefused(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	made := library(t, src, &syncer{})

	cases := map[string][]byte{
		"a text file pretending": []byte("this is not a picture at all, it is just words"),
		"something empty":        {},
		"a program":              {0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00},
	}

	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := made.Add(context.Background(), anEditor(project), project, "thing.png", content); !errors.Is(err, httpapi.ErrRefused) {
				t.Errorf("Add() error = %v, want it refused", err)
			}
		})
	}
	if len(src.site.saved) != 0 {
		t.Error("something was written anyway")
	}
}

func TestAPictureTooLargeIsRefusedWithItsSizeInWords(t *testing.T) {
	project := uuid.New()
	made := library(t, &sources{site: aSite(t)}, &syncer{})

	huge := make([]byte, media.MaxBytes+1)
	copy(huge, aPNG(t, 10, 10))

	_, err := made.Add(context.Background(), anEditor(project), project, "huge.png", huge)
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Add() error = %v, want it refused", err)
	}
	if !strings.Contains(err.Error(), "MB") {
		t.Errorf("error = %q, want it to say how big it is in words", err)
	}
}

func TestADrawingIsTakenAsAPicture(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}

	drawing := []byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><circle r="5"/></svg>`)
	added, err := library(t, src, &syncer{}).Add(context.Background(), anEditor(project), project,
		"circle.svg", drawing)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if added.Name != "circle.svg" {
		t.Errorf("Name = %q", added.Name)
	}
}

func TestADescriptionIsWrittenIntoTheSite(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	sync := &syncer{}

	described, err := library(t, src, sync).Describe(context.Background(), anEditor(project), project,
		"logo.png", "The bakery's sign above the door")
	if err != nil {
		t.Fatalf("Describe() error = %v", err)
	}

	if described.Alt != "The bakery's sign above the door" {
		t.Errorf("Alt = %q", described.Alt)
	}
	written := string(src.site.saved[0].Writes[media.DescriptionsFile])
	if !strings.Contains(written, "logo.png") || !strings.Contains(written, "above the door") {
		t.Errorf("the descriptions file holds:\n%s", written)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times", sync.asked)
	}
}

func TestDescriptionsAreWrittenInTheSameOrderEveryTime(t *testing.T) {
	first, err := media.WriteDescriptions(map[string]string{"b.png": "bee", "a.png": "ay", "c.png": "sea"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := media.WriteDescriptions(map[string]string{"c.png": "sea", "a.png": "ay", "b.png": "bee"})
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(again) {
		t.Errorf("the same descriptions made two different files:\n%s\n%s", first, again)
	}
}

func TestAnEmptyDescriptionIsNotKept(t *testing.T) {
	written, err := media.WriteDescriptions(map[string]string{"a.png": "   ", "b.png": "bee"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(written), "a.png") {
		t.Errorf("an empty description was kept:\n%s", written)
	}
}

func TestAPictureNothingUsesIsRemoved(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	sync := &syncer{}

	if err := library(t, src, sync).Remove(context.Background(), anEditor(project), project, "spare.png", false); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	save := src.site.saved[0]
	if len(save.Deletes) != 1 || save.Deletes[0] != "static/images/spare.png" {
		t.Errorf("deleted %v", save.Deletes)
	}
	if sync.asked != 1 {
		t.Errorf("asked for a sync %d times", sync.asked)
	}
}

func TestAPictureStillInUseIsNotRemovedWithoutSaying(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	made := library(t, src, &syncer{})

	err := made.Remove(context.Background(), anEditor(project), project, "logo.png", false)
	if !errors.Is(err, httpapi.ErrRefused) {
		t.Fatalf("Remove() error = %v, want it refused", err)
	}
	if !strings.Contains(err.Error(), "still using it") {
		t.Errorf("error = %q, want it to say why", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("it was removed anyway")
	}

	if err := made.Remove(context.Background(), anEditor(project), project, "logo.png", true); err != nil {
		t.Fatalf("Remove() when told to anyway error = %v", err)
	}
	if len(src.site.saved) != 1 {
		t.Error("it was not removed when asked for a second time")
	}
}

func TestRemovingAPictureTakesItsDescriptionToo(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	made := library(t, src, &syncer{})

	if _, err := made.Describe(context.Background(), anEditor(project), project, "spare.png", "a spare"); err != nil {
		t.Fatal(err)
	}
	if err := made.Remove(context.Background(), anEditor(project), project, "spare.png", false); err != nil {
		t.Fatal(err)
	}

	written := string(src.site.files[media.DescriptionsFile].Data)
	if strings.Contains(written, "spare.png") {
		t.Errorf("the description outlived its picture:\n%s", written)
	}
}

func TestAPictureIsReadBackSmallOrWhole(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	made := library(t, src, &syncer{})

	whole, kind, err := made.Read(context.Background(), anEditor(project), project, "logo.png", false)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if kind != "image/png" || len(whole) == 0 {
		t.Errorf("kind = %q, %d bytes", kind, len(whole))
	}

	big := aPNG(t, 1200, 600)
	src.site.files["static/images/big.png"] = &fstest.MapFile{Data: big}

	small, kind, err := made.Read(context.Background(), anEditor(project), project, "big.png", true)
	if err != nil {
		t.Fatalf("Read() small error = %v", err)
	}
	if kind != "image/png" {
		t.Errorf("kind = %q", kind)
	}
	if len(small) >= len(big) {
		t.Errorf("the small copy is %d bytes and the picture is %d; want it smaller", len(small), len(big))
	}

	decoded, err := png.Decode(bytes.NewReader(small))
	if err != nil {
		t.Fatalf("the small copy is not a picture: %v", err)
	}
	if decoded.Bounds().Dx() != media.ThumbnailWidth {
		t.Errorf("the small copy is %d wide, want %d", decoded.Bounds().Dx(), media.ThumbnailWidth)
	}
}

func TestASmallCopyIsMadeOnceAndKept(t *testing.T) {
	dir := t.TempDir()
	thumbnails, err := media.NewThumbnails(dir)
	if err != nil {
		t.Fatal(err)
	}

	picture := aPNG(t, 900, 400)
	first, _, err := thumbnails.Of(picture, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	again, _, err := thumbnails.Of(picture, "image/png")
	if err != nil {
		t.Fatal(err)
	}

	if string(first) != string(again) {
		t.Error("asking twice gave two different copies")
	}

	// The cache is named after the bytes, so one picture is one file and
	// throwing the directory away costs only the work of making it again.
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 1 {
		t.Errorf("the cache holds %d files, want one for one picture", len(entries))
	}
}

func TestAPictureThatIsNotThereIsNotFound(t *testing.T) {
	project := uuid.New()
	made := library(t, &sources{site: aSite(t)}, &syncer{})

	for _, name := range []string{"absent.png", "", "..", "../../etc/passwd", "a/b.png"} {
		if _, _, err := made.Read(context.Background(), anEditor(project), project, name, false); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("Read(%q) error = %v, want ErrMissing", name, err)
		}
		if err := made.Remove(context.Background(), anEditor(project), project, name, true); !errors.Is(err, httpapi.ErrMissing) {
			t.Errorf("Remove(%q) error = %v, want ErrMissing", name, err)
		}
	}
}

func TestAViewerMayLookAndChangeNothing(t *testing.T) {
	project := uuid.New()
	src := &sources{site: aSite(t)}
	made := library(t, src, &syncer{})

	if _, err := made.List(context.Background(), aViewer(project), project); err != nil {
		t.Errorf("List() as a viewer error = %v, want it allowed", err)
	}
	if _, _, err := made.Read(context.Background(), aViewer(project), project, "logo.png", true); err != nil {
		t.Errorf("Read() as a viewer error = %v, want it allowed", err)
	}
	if _, err := made.Add(context.Background(), aViewer(project), project, "x.png", aPNG(t, 10, 10)); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Add() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if err := made.Remove(context.Background(), aViewer(project), project, "spare.png", true); !errors.Is(err, httpapi.ErrNotAllowed) {
		t.Errorf("Remove() as a viewer error = %v, want ErrNotAllowed", err)
	}
	if len(src.site.saved) != 0 {
		t.Error("a viewer changed the site")
	}
}

func TestASiteKeepsTheFolderItAlreadyUses(t *testing.T) {
	already := fstest.MapFS{
		"static/uploads/logo.png": {Data: []byte("x")},
	}
	if folder := media.FolderIn(already); folder != "static/uploads" {
		t.Errorf("FolderIn() = %q, want the folder the site already uses", folder)
	}

	if folder := media.FolderIn(fstest.MapFS{}); folder != media.Folder {
		t.Errorf("FolderIn() on a new site = %q, want %q", folder, media.Folder)
	}
}

func TestANameIsTidiedTheWayEverythingElseIs(t *testing.T) {
	cases := []struct{ said, extension, want string }{
		{"Our New Shop.PNG", ".png", "our-new-shop.png"},
		{"foto 2026.jpg", ".jpg", "foto-2026.jpg"},
		{"a/b/c.png", ".png", "a-b-c.png"},
		{"...png", ".png", ""},
		{"", ".png", ""},
		// The extension comes from what the bytes are, not from what the
		// file was called, so a picture misnamed by its sender is still
		// named for what it is.
		{"photo.png", ".jpg", "photo.jpg"},
	}

	for _, test := range cases {
		got, err := media.Name(test.said, test.extension)
		if test.want == "" {
			if err == nil {
				t.Errorf("Name(%q) = %q, want it refused", test.said, got)
			}
			continue
		}
		if err != nil || got != test.want {
			t.Errorf("Name(%q, %q) = %q, %v, want %q", test.said, test.extension, got, err, test.want)
		}
	}
}
