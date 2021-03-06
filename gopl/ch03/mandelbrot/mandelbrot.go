package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/cmplx"
	"os"
	"sync"

	"github.com/krasoffski/gomill/htcmap"
)

const (
	xmin, ymin = -2.2, -1.2
	xmax, ymax = +1.2, +1.2
)

type point struct {
	x, y int
}

type pixel struct {
	point
	c color.Color
}

func xCord(x, width, factor int) float64 {
	return float64(x)/float64(width*factor)*(xmax-xmin) + xmin
}

func yCord(y, height, factor int) float64 {
	return float64(y)/float64(height*factor)*(ymax-ymin) + ymin
}

func superSampling(p *point, width, height, factor int, smooth bool) color.Color {

	xCords, yCords := make([]float64, factor), make([]float64, factor)
	subPixels := make([]color.Color, factor*factor)

	// Single calculation of required coordinates for super sampling.
	for i := 0; i < factor; i++ {
		xCords[i] = xCord(p.x+i, width, factor)
		yCords[i] = yCord(p.y+i, height, factor)
	}

	// Instead of calculation coordinate only fetching required one.
	for iy := 0; iy < factor; iy++ {
		for ix := 0; ix < factor; ix++ {
			// Using one dimension array because do not care about pixel order,
			// because at the end we are calculating avarage for all sub-pixels.
			subPixels[iy*factor+ix] = mandelbrot(
				complex(xCords[ix], yCords[iy]), smooth)
		}
	}

	var rAvg, gAvg, bAvg float64

	// TODO: think about removing multiplication of factor for each calculation.
	factor2 := float64(factor * factor)
	for _, c := range subPixels {
		r, g, b, _ := c.RGBA()
		rAvg += float64(r) / factor2
		gAvg += float64(g) / factor2
		bAvg += float64(b) / factor2
	}
	return color.RGBA64{uint16(rAvg), uint16(gAvg), uint16(bAvg), 0xFFFF}
}

func mandelbrot(z complex128, smooth bool) color.Color {
	const iterations = 255
	const contrast = 15

	var v complex128
	var step float64
	var skip uint8

	// TODO: create more common approach here
	if smooth {
		skip = 2
	}

	for n := uint8(0); n < iterations; n++ {
		v = v*v + z
		vAbs := cmplx.Abs(v)
		if vAbs > 2 && n > skip {
			if smooth {
				step = (float64(n) + 1 - math.Log(math.Log(vAbs))/math.Log(2)) * contrast
			} else {
				step = float64(n * contrast)
			}
			r, g, b := htcmap.AsUInt8(step, 0, iterations)
			return color.RGBA{r, g, b, 255}
		}
	}
	return color.Black
}

func compute(width, height, factor, workers int, smooth bool) <-chan *pixel {
	var wg sync.WaitGroup
	points := make(chan *point)
	pixels := make(chan *pixel, workers)

	go func() {
		defer close(points)

		for py := 0; py < height*factor; py += factor {
			for px := 0; px < width*factor; px += factor {
				points <- &point{px, py}
			}
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				p, ok := <-points
				if !ok {
					return
				}
				c := superSampling(p, width, height, factor, smooth)
				pixels <- &pixel{point{p.x / factor, p.y / factor}, c}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(pixels)
	}()

	return pixels
}

func main() {
	factor := flag.Int("factor", 2, "scale factor for super sampling, [1, 10]")
	workers := flag.Int("workers", 2, "number of workers for calculation, [1, 512]")
	width := flag.Int("width", 1536, "width of png image in pixels, WxH=3x2")
	height := flag.Int("height", 1024, "height of png image in pixels, WxH=3x2")
	smooth := flag.Bool("smooth", false, "enables smooth color transition")
	flag.Parse()
	if *factor < 1 || *factor > 10 {
		fmt.Fprintf(os.Stderr, "error: invalid value '%d', [1, 10]\n", *factor)
		os.Exit(1)
	}

	if *workers < 1 || *workers > 512 {
		fmt.Fprintf(os.Stderr, "error: invalid value '%d', [1, 255]\n", *workers)
		os.Exit(1)
	}

	img := image.NewRGBA(image.Rect(0, 0, *width, *height))

	for p := range compute(*width, *height, *factor, *workers, *smooth) {
		img.Set(p.x, p.y, p.c)
	}

	if err := png.Encode(os.Stdout, img); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding png: %s", err)
		os.Exit(1)
	}
	if err := png.Encode(os.Stdout, img); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding png: %s", err)
		os.Exit(1)
	}
}
