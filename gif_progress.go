package gif_progress

import (
	"image"
	"image/color"
	"image/gif"
	"math"

	colorful "github.com/lucasb-eyer/go-colorful"
)

func AddProgressBar(inGif *gif.GIF, barTop bool, barHeight int, barColor color.RGBA) *gif.GIF {
	// NOTE: inGif is changed destructively
	width := inGif.Config.Width
	height := inGif.Config.Height
	image_len := len(inGif.Image)
	for i, paletted := range inGif.Image {
		w := int(float32(width) * ((float32(i) + 1) / float32(image_len)))
		for x := 0; x < w; x++ {
			for h := 0; h < barHeight; h++ {
				var y = h
				if !barTop {
					y = height - h
				}
				paletted.Set(x, y, barColor)
			}
		}
	}
	return inGif
}

func AddProgressBarFPS(inGif *gif.GIF, barTop bool, barHeight int, barColor color.RGBA, barFPS int) *gif.GIF {
	// Preparing GIF instance
	totalLength := 0
	for _, delay := range inGif.Delay {
		totalLength += delay
	}
	totalLengthSec := float32(totalLength) / float32(100)
	frameCount := int(totalLengthSec * float32(barFPS))
	outGif := &gif.GIF{
		Image:     make([]*image.Paletted, 0, frameCount+5), // NOTE: +5 is for rounding error
		Delay:     make([]int, 0, frameCount+5),             // NOTE: +5 is for rounding error
		LoopCount: inGif.LoopCount,
		Disposal:  make([]byte, 0, frameCount+5), // NOTE: +5 is for rounding error
		Config: image.Config{
			Width:      inGif.Config.Width,
			Height:     inGif.Config.Height,
			ColorModel: nil,
		},
		BackgroundIndex: inGif.BackgroundIndex,
	}

	// Frame inserting function
	insertFrame := func(frameIndex int, delay int, currentTime float32) {
		// Modify original frame if the rectangle is overlapped with progressbar
		w := int(float32(inGif.Config.Width) * (currentTime / totalLengthSec))
		if w > inGif.Config.Width {
			w = inGif.Config.Width
		}
		h := barHeight
		zeroH := 0
		if !barTop {
			h = inGif.Config.Height - barHeight
			zeroH = inGif.Config.Height
		}
		if (image.Point{0, zeroH}).In(inGif.Image[frameIndex].Rect) || (image.Point{w, zeroH}).In(inGif.Image[frameIndex].Rect) || (image.Point{0, h}).In(inGif.Image[frameIndex].Rect) || (image.Point{w, h}).In(inGif.Image[frameIndex].Rect) {
			// Create new frame from original to modify
			newFrame := &image.Paletted{
				Pix:     make([]uint8, len(inGif.Image[frameIndex].Pix)),
				Stride:  inGif.Image[frameIndex].Stride,
				Rect:    inGif.Image[frameIndex].Rect,
				Palette: make([]color.Color, len(inGif.Image[frameIndex].Palette)),
			}
			copy(newFrame.Pix, inGif.Image[frameIndex].Pix)
			copy(newFrame.Palette, inGif.Image[frameIndex].Palette)

			// Add barColor to the palette
			if len(newFrame.Palette) >= 255 {
				var colorIdx uint8
				var colorConvIdx uint8
				distance := math.MaxFloat32
				for i, pickedA := range newFrame.Palette {
					for j, pickedB := range newFrame.Palette {
						if i != j {
							pickedColorfulA, _ := colorful.MakeColor(pickedA)
							pickedColorfulB, _ := colorful.MakeColor(pickedB)
							pickedDistance := pickedColorfulA.DistanceHPLuv(pickedColorfulB)
							if pickedDistance < distance {
								distance = pickedDistance
								colorIdx = uint8(j)
								colorConvIdx = uint8(i)
							}
						}
					}
				}
				for pixelIdx, pixel := range newFrame.Pix {
					if pixel == colorIdx {
						newFrame.Pix[pixelIdx] = colorConvIdx
					}
				}
				newPalette := append(newFrame.Palette[:colorIdx], barColor)
				if colorIdx < 255 {
					newPalette = append(newPalette, newFrame.Palette[colorIdx+1:]...)
				}
				newFrame.Palette = newPalette
			} else {
				newFrame.Palette = append(newFrame.Palette, barColor)
			}

			// Draw progressbar
			for x := 0; x < w; x++ {
				for h := 0; h < barHeight; h++ {
					var y = h
					if !barTop {
						y = inGif.Config.Height - h
					}
					newFrame.Set(x, y, barColor)
				}
			}
			outGif.Delay = append(outGif.Delay, delay)
			outGif.Image = append(outGif.Image, newFrame)
			outGif.Disposal = append(outGif.Disposal, inGif.Disposal[frameIndex])
		} else {
			outGif.Delay = append(outGif.Delay, delay)
			outGif.Image = append(outGif.Image, inGif.Image[frameIndex])
			outGif.Disposal = append(outGif.Disposal, inGif.Disposal[frameIndex])
		}

		// Add a progressbar-only frame
		if w > 0 {
			startHeight := 0
			endHeight := barHeight - 1
			if !barTop {
				startHeight = inGif.Config.Height - barHeight + 1
				endHeight = inGif.Config.Height
			}
			newPFrame := &image.Paletted{
				Pix:    make([]uint8, 1200*500),
				Stride: 0,
				Rect: image.Rectangle{
					Min: image.Point{0, startHeight},
					Max: image.Point{w, endHeight},
				},
				Palette: []color.Color{barColor},
			}
			outGif.Image = append(outGif.Image, newPFrame)
			outGif.Delay = append(outGif.Delay, 0)
			outGif.Disposal = append(outGif.Disposal, inGif.Disposal[frameIndex])
		}
	}

	// Insert frames
	commonDelay := int(1 / float32(barFPS) * 100)
	commonTimeCounter := float32(0)
	for j := 0; j < len(inGif.Image); j++ {
		addingFrameCount := inGif.Delay[j] / commonDelay
		remainingFrameDelay := inGif.Delay[j] % commonDelay
		for i := 0; i < addingFrameCount; i++ {
			insertFrame(j, commonDelay, commonTimeCounter)
			commonTimeCounter += float32(commonDelay) / float32(100)
		}
		insertFrame(j, remainingFrameDelay, commonTimeCounter)
		commonTimeCounter += float32(remainingFrameDelay) / float32(100)
	}
	return outGif
}
