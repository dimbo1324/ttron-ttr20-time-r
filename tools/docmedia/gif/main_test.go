package main

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// frame paints a background and one square of a second colour, so a pair of
// frames differs in a region whose bounds are known exactly.
func frame(w, h int, background color.RGBA, square image.Rectangle, mark color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetRGBA(x, y, background)
		}
	}
	for y := square.Min.Y; y < square.Max.Y; y++ {
		for x := square.Min.X; x < square.Max.X; x++ {
			img.SetRGBA(x, y, mark)
		}
	}
	return img
}

func writeFrames(t *testing.T, dir string, images ...image.Image) {
	t.Helper()
	for i, img := range images {
		file, err := os.Create(filepath.Join(dir, string(rune('a'+i))+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(file, img); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

var (
	dark  = color.RGBA{12, 14, 18, 0xff}
	teal  = color.RGBA{0, 190, 200, 0xff}
	amber = color.RGBA{220, 170, 40, 0xff}
)

func TestRunWritesOneFrameForEachChange(t *testing.T) {
	in := t.TempDir()
	writeFrames(t, in,
		frame(40, 30, dark, image.Rect(2, 2, 8, 8), teal),
		frame(40, 30, dark, image.Rect(2, 2, 8, 8), amber),
		frame(40, 30, dark, image.Rect(20, 20, 30, 28), teal),
	)
	out := filepath.Join(t.TempDir(), "nested", "out.gif")

	if err := run(in, out, 10, 0, 0); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	anim, err := gif.DecodeAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(anim.Image) != 3 {
		t.Fatalf("frames = %d, want 3", len(anim.Image))
	}

	// The first frame is whole; the rest are only what moved. The third frame
	// changes two squares -- the one being erased and the one being drawn --
	// so its bounds cover both.
	if got, want := anim.Image[0].Bounds(), image.Rect(0, 0, 40, 30); got != want {
		t.Errorf("first frame bounds = %v, want the whole image %v", got, want)
	}
	if got, want := anim.Image[1].Bounds(), image.Rect(2, 2, 8, 8); got != want {
		t.Errorf("second frame bounds = %v, want only the square that changed %v", got, want)
	}
	if got := anim.Image[2].Bounds(); got.Dx() >= 40 && got.Dy() >= 30 {
		t.Errorf("third frame bounds = %v, want a sub-rectangle", got)
	}
}

func TestRunHoldsAFrameThatDidNotChange(t *testing.T) {
	in := t.TempDir()
	still := frame(20, 20, dark, image.Rect(4, 4, 10, 10), teal)
	writeFrames(t, in, still, still, still)
	out := filepath.Join(t.TempDir(), "still.gif")

	if err := run(in, out, 10, 0, 0); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	anim, err := gif.DecodeAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(anim.Image) != 1 {
		t.Fatalf("frames = %d, want 1: identical frames should be held, not repeated", len(anim.Image))
	}
	// The two frames that were folded away are paid back as delay, so the
	// animation still lasts as long as it was asked to.
	if anim.Delay[0] != 30 {
		t.Errorf("delay = %d, want 30 (the held frames' time)", anim.Delay[0])
	}
}

func TestRunRejectsAnEmptyDirectory(t *testing.T) {
	if err := run(t.TempDir(), filepath.Join(t.TempDir(), "x.gif"), 10, 0, 0); err == nil {
		t.Fatal("expected an error for a directory with no frames")
	}
}

func TestFinalDelayOverridesTheLastFrame(t *testing.T) {
	in := t.TempDir()
	writeFrames(t, in,
		frame(20, 20, dark, image.Rect(1, 1, 5, 5), teal),
		frame(20, 20, dark, image.Rect(1, 1, 5, 5), amber),
	)
	out := filepath.Join(t.TempDir(), "final.gif")

	if err := run(in, out, 10, 250, 0); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	anim, err := gif.DecodeAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if last := anim.Delay[len(anim.Delay)-1]; last != 250 {
		t.Errorf("final delay = %d, want 250", last)
	}
}

func TestChangedFindsNothingInIdenticalFrames(t *testing.T) {
	palette := color.Palette{color.Black, color.White}
	a := image.NewPaletted(image.Rect(0, 0, 8, 8), palette)
	b := image.NewPaletted(image.Rect(0, 0, 8, 8), palette)
	if got := changed(a, b); !got.Empty() {
		t.Errorf("changed = %v, want an empty rectangle", got)
	}

	b.SetColorIndex(3, 5, 1)
	if got, want := changed(a, b), image.Rect(3, 5, 4, 6); got != want {
		t.Errorf("changed = %v, want %v", got, want)
	}
}

func TestQuantizeSpendsThePaletteOnColoursThatAreThere(t *testing.T) {
	// Three colours, and room for 256. Every one of them should survive
	// exactly rather than be approximated by a lattice point nearby.
	frames := []image.Image{frame(30, 30, dark, image.Rect(5, 5, 15, 15), teal)}
	palette := quantize(frames, 256)

	for _, want := range []color.RGBA{dark, teal} {
		got := palette.Convert(want).(color.RGBA)
		if got.R != want.R || got.G != want.G || got.B != want.B {
			t.Errorf("nearest palette entry to %v is %v, want an exact match", want, got)
		}
	}
}

func TestQuantizeSurvivesAnImageWithNoPixels(t *testing.T) {
	if got := quantize(nil, 256); len(got) == 0 {
		t.Fatal("quantize returned an empty palette")
	}
}

func TestReadFramesReportsAFileItCannotDecode(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.png"), []byte("not a png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readFrames(dir); err == nil {
		t.Fatal("expected an error naming the file that would not decode")
	}
}
