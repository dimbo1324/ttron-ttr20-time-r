// Command gif turns a directory of PNG frames into one animated GIF.
//
// The documentation shows the console moving, and a repository is a bad place
// for a video: a `.mp4` committed to a tree does not play inside a README on
// GitHub, it downloads. An animated GIF plays everywhere Markdown is rendered,
// which is the only property that matters here.
//
// ffmpeg would do this in one line and is not assumed to be installed, so the
// encoder is the standard library's, and the two things the standard library
// leaves to the caller -- choosing 256 colours, and not re-encoding the
// unchanged nine tenths of every frame -- are done below.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	in := flag.String("in", "", "directory of PNG frames, encoded in filename order")
	out := flag.String("out", "", "animated GIF to write")
	delay := flag.Int("delay", 12, "delay between frames, in hundredths of a second")
	loop := flag.Int("loop", 0, "loop count; 0 repeats forever")
	final := flag.Int("final-delay", 0, "delay on the last frame, in hundredths of a second (0 uses -delay)")
	flag.Parse()

	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: gif -in <frames dir> -out <file.gif> [-delay 12]")
		os.Exit(2)
	}
	if err := run(*in, *out, *delay, *final, *loop); err != nil {
		fmt.Fprintln(os.Stderr, "gif:", err)
		os.Exit(1)
	}
}

func run(in, out string, delay, final, loop int) error {
	frames, err := readFrames(in)
	if err != nil {
		return err
	}
	if len(frames) == 0 {
		return fmt.Errorf("no .png frames in %s", in)
	}

	// One palette for the whole animation rather than one per frame. A palette
	// that changes between frames makes a static background shimmer, which on
	// a screen recording of a mostly-still page is the only thing the eye
	// follows.
	palette := quantize(frames, 256)

	anim := gif.GIF{LoopCount: loop}
	var previous *image.Paletted

	for _, frame := range frames {
		full := image.NewPaletted(frame.Bounds(), palette)
		draw.FloydSteinberg.Draw(full, frame.Bounds(), frame, image.Point{})

		region := frame.Bounds()
		if previous != nil {
			region = changed(previous, full)
			if region.Empty() {
				// Nothing moved: hold the frame already on screen instead of
				// writing a duplicate of it.
				anim.Delay[len(anim.Delay)-1] += delay
				continue
			}
		}

		anim.Image = append(anim.Image, full.SubImage(region).(*image.Paletted))
		anim.Delay = append(anim.Delay, delay)
		anim.Disposal = append(anim.Disposal, gif.DisposalNone)
		previous = full
	}

	if final > 0 && len(anim.Delay) > 0 {
		anim.Delay[len(anim.Delay)-1] = final
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	file, err := os.Create(out)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if err := gif.EncodeAll(file, &anim); err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		return err
	}
	fmt.Printf("%s: %d frames, %d colours, %.1f KiB\n",
		filepath.ToSlash(out), len(anim.Image), len(palette), float64(info.Size())/1024)
	return nil
}

func readFrames(dir string) ([]image.Image, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.png"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	frames := make([]image.Image, 0, len(names))
	for _, name := range names {
		file, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		img, err := png.Decode(file)
		_ = file.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		frames = append(frames, img)
	}
	return frames, nil
}

// changed returns the smallest rectangle covering every pixel that differs
// between two frames.
//
// A console page is nearly all background: one number ticks and a row appears
// in a log. Writing only the rectangle that moved is the difference between a
// GIF of a few hundred kilobytes and one of several megabytes.
func changed(previous, current *image.Paletted) image.Rectangle {
	bounds := current.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if previous.ColorIndexAt(x, y) == current.ColorIndexAt(x, y) {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if minX > maxX {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// ---------------------------------------------------------------- quantizing

// quantize picks a palette by median cut: the colours of every frame are put
// in one box, the box is split repeatedly along its longest channel at the
// median, and each final box contributes its average.
//
// The stock `palette.Plan9` is an even lattice over the whole cube, which
// spends most of its entries on colours a dark interface never uses and bands
// the handful of near-black greys it is almost entirely made of.
func quantize(frames []image.Image, size int) color.Palette {
	boxes := []colorBox{{pixels: sample(frames)}}
	if len(boxes[0].pixels) == 0 {
		return color.Palette{color.Black, color.White}
	}
	boxes[0].fit()

	for len(boxes) < size {
		// Split whichever box still spans the most colour. Splitting the
		// largest by pixel count instead would spend the palette on the
		// background, which is one colour however many pixels it covers.
		widest, at := 0, -1
		for i := range boxes {
			if span := boxes[i].span(); span > widest && len(boxes[i].pixels) > 1 {
				widest, at = span, i
			}
		}
		if at < 0 {
			break
		}
		a, b := boxes[at].split()
		boxes[at] = a
		boxes = append(boxes, b)
	}

	out := make(color.Palette, 0, len(boxes))
	for i := range boxes {
		out = append(out, boxes[i].average())
	}
	return out
}

type rgb struct{ r, g, b uint8 }

type colorBox struct {
	pixels   []rgb
	min, max rgb
}

// sample walks the frames on a stride rather than reading every pixel. A
// 1280x720 frame is nearly a million pixels and a palette does not get better
// for having counted all of them.
func sample(frames []image.Image) []rgb {
	const stride = 3

	var pixels []rgb
	for _, frame := range frames {
		bounds := frame.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y += stride {
			for x := bounds.Min.X; x < bounds.Max.X; x += stride {
				r, g, b, _ := frame.At(x, y).RGBA()
				pixels = append(pixels, rgb{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)})
			}
		}
	}
	return pixels
}

func (c *colorBox) fit() {
	c.min = rgb{255, 255, 255}
	c.max = rgb{0, 0, 0}
	for _, p := range c.pixels {
		c.min.r = min(c.min.r, p.r)
		c.min.g = min(c.min.g, p.g)
		c.min.b = min(c.min.b, p.b)
		c.max.r = max(c.max.r, p.r)
		c.max.g = max(c.max.g, p.g)
		c.max.b = max(c.max.b, p.b)
	}
}

func (c *colorBox) span() int {
	return max(int(c.max.r-c.min.r), int(c.max.g-c.min.g), int(c.max.b-c.min.b))
}

func (c *colorBox) split() (colorBox, colorBox) {
	dr := int(c.max.r - c.min.r)
	dg := int(c.max.g - c.min.g)
	db := int(c.max.b - c.min.b)

	channel := func(p rgb) uint8 { return p.r }
	switch {
	case dg >= dr && dg >= db:
		channel = func(p rgb) uint8 { return p.g }
	case db >= dr && db >= dg:
		channel = func(p rgb) uint8 { return p.b }
	}
	sort.Slice(c.pixels, func(i, j int) bool { return channel(c.pixels[i]) < channel(c.pixels[j]) })

	half := len(c.pixels) / 2
	a := colorBox{pixels: c.pixels[:half]}
	b := colorBox{pixels: c.pixels[half:]}
	a.fit()
	b.fit()
	return a, b
}

func (c *colorBox) average() color.Color {
	var r, g, b int
	for _, p := range c.pixels {
		r += int(p.r)
		g += int(p.g)
		b += int(p.b)
	}
	n := len(c.pixels)
	if n == 0 {
		return color.Black
	}
	return color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), 0xff}
}
