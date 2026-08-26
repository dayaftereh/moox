package moo2gfx

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

const (
	FlagJunction        uint16 = 0x2000
	FlagInternalPalette uint16 = 0x1000
	FlagFunctionalColor uint16 = 0x0800
	FlagFillBackground  uint16 = 0x0400
	FlagNoCompression   uint16 = 0x0100
)

type Graphic struct {
	Width        int
	Height       int
	FrameCount   int
	Delay        uint16
	Flags        uint16
	FrameOffsets []uint32
	PaletteShift int
	Palette      [256]color.NRGBA
	PaletteValid [256]bool
	Data         []byte
}

type DecodeInfo struct {
	MissingPalettePixels int
	DrawnPixels          int
}

type DisplayDecoder struct {
	graphic   *Graphic
	next      int
	composite *image.NRGBA
	broken    bool
}

func (g *Graphic) NewDisplayDecoder() *DisplayDecoder {
	return &DisplayDecoder{graphic: g}
}

func (d *DisplayDecoder) DecodeNext() (*image.NRGBA, DecodeInfo, error) {
	var info DecodeInfo
	if d.next >= d.graphic.FrameCount {
		return nil, info, fmt.Errorf("display frame %d out of range [0,%d)", d.next, d.graphic.FrameCount)
	}
	frame := d.next
	d.next++
	if d.broken {
		return nil, info, fmt.Errorf("junction display sequence is incomplete after an earlier frame decode failure")
	}

	delta, info, err := d.graphic.DecodeFrame(frame)
	if err != nil {
		if d.graphic.Flags&FlagJunction != 0 {
			d.broken = true
		}
		return nil, info, err
	}
	if d.graphic.Flags&FlagJunction == 0 {
		return delta, info, nil
	}
	if d.composite == nil {
		d.composite = image.NewNRGBA(image.Rect(0, 0, d.graphic.Width, d.graphic.Height))
	}
	draw.Draw(d.composite, d.composite.Bounds(), delta, image.Point{}, draw.Over)
	return d.composite, info, nil
}
func Parse(data []byte) (*Graphic, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("graphic payload too small: %d bytes", len(data))
	}
	width := int(binary.LittleEndian.Uint16(data[0:2]))
	height := int(binary.LittleEndian.Uint16(data[2:4]))
	reserved := binary.LittleEndian.Uint16(data[4:6])
	frames := int(binary.LittleEndian.Uint16(data[6:8]))
	delay := binary.LittleEndian.Uint16(data[8:10])
	flags := binary.LittleEndian.Uint16(data[10:12])
	if width <= 0 || height <= 0 || width > 4096 || height > 4096 || reserved != 0 || frames <= 0 || frames > 4096 {
		return nil, fmt.Errorf("not a plausible MOO2 graphic header: %dx%d reserved=%d frames=%d", width, height, reserved, frames)
	}
	offsetBytes := (frames + 1) * 4
	headerEnd := 12 + offsetBytes
	if headerEnd > len(data) {
		return nil, fmt.Errorf("graphic frame-offset table exceeds payload")
	}
	offsets := make([]uint32, frames+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[12+i*4 : 16+i*4])
		if int(offsets[i]) > len(data) {
			return nil, fmt.Errorf("frame offset %d exceeds payload: %d > %d", i, offsets[i], len(data))
		}
		if i > 0 && offsets[i] < offsets[i-1] {
			return nil, fmt.Errorf("frame offsets are not monotonic at %d", i)
		}
	}

	g := &Graphic{Width: width, Height: height, FrameCount: frames, Delay: delay, Flags: flags, FrameOffsets: offsets, Data: data}
	if flags&FlagInternalPalette != 0 {
		if headerEnd+4 > len(data) {
			return nil, fmt.Errorf("internal palette header missing")
		}
		shift := int(binary.LittleEndian.Uint16(data[headerEnd : headerEnd+2]))
		count := int(binary.LittleEndian.Uint16(data[headerEnd+2 : headerEnd+4]))
		if shift < 0 || count < 0 || shift+count > 256 {
			return nil, fmt.Errorf("invalid internal palette shift=%d count=%d", shift, count)
		}
		paletteEnd := headerEnd + 4 + count*4
		if paletteEnd > len(data) {
			return nil, fmt.Errorf("internal palette exceeds payload")
		}
		if len(offsets) > 0 && uint32(paletteEnd) > offsets[0] {
			return nil, fmt.Errorf("internal palette overlaps first frame")
		}
		g.PaletteShift = shift
		for i := 0; i < count; i++ {
			p := headerEnd + 4 + i*4
			r := scaleDAC(data[p+1])
			gg := scaleDAC(data[p+2])
			b := scaleDAC(data[p+3])
			idx := shift + i
			g.Palette[idx] = color.NRGBA{R: r, G: gg, B: b, A: 255}
			g.PaletteValid[idx] = true
		}
	}
	return g, nil
}

func scaleDAC(v byte) byte {
	n := int(v) * 4
	if n > 255 {
		n = 255
	}
	return byte(n)
}

func (g *Graphic) DecodeFrame(frame int) (*image.NRGBA, DecodeInfo, error) {
	var info DecodeInfo
	if frame < 0 || frame >= g.FrameCount {
		return nil, info, fmt.Errorf("frame %d out of range [0,%d)", frame, g.FrameCount)
	}
	start := int(g.FrameOffsets[frame])
	end := int(g.FrameOffsets[frame+1])
	if start+4 > end || end > len(g.Data) {
		return nil, info, fmt.Errorf("invalid frame %d range [%d,%d)", frame, start, end)
	}
	if g.Flags&FlagNoCompression != 0 {
		return g.decodeUncompressedFrame(frame, start, end)
	}

	indicator := binary.LittleEndian.Uint16(g.Data[start : start+2])
	if indicator != 1 {
		return nil, info, fmt.Errorf("frame %d has unsupported start indicator %d", frame, indicator)
	}
	y := int(binary.LittleEndian.Uint16(g.Data[start+2 : start+4]))
	x := 0
	pos := start + 4
	img := image.NewNRGBA(image.Rect(0, 0, g.Width, g.Height))
	maxSequences := g.Width*g.Height*2 + 4096
	sequences := 0

	for pos < end {
		sequences++
		if sequences > maxSequences {
			return nil, info, fmt.Errorf("frame %d exceeded sequence safety limit", frame)
		}
		if pos+2 > end {
			return nil, info, fmt.Errorf("truncated sequence count in frame %d", frame)
		}
		count := int(binary.LittleEndian.Uint16(g.Data[pos : pos+2]))
		pos += 2
		if count == 0 {
			if pos+2 > end {
				return nil, info, fmt.Errorf("truncated y-indent in frame %d", frame)
			}
			dy := int(binary.LittleEndian.Uint16(g.Data[pos : pos+2]))
			pos += 2
			if dy == 1000 {
				return img, info, nil
			}
			y += dy
			x = 0
			continue
		}

		if pos+2 > end {
			return nil, info, fmt.Errorf("truncated x-indent in frame %d", frame)
		}
		dx := int(binary.LittleEndian.Uint16(g.Data[pos : pos+2]))
		pos += 2
		x += dx
		if pos+count > end {
			return nil, info, fmt.Errorf("pixel run exceeds frame %d payload", frame)
		}
		for i := 0; i < count; i++ {
			px := x + i
			idx := int(g.Data[pos+i])
			if px < 0 || px >= g.Width || y < 0 || y >= g.Height {
				continue
			}
			if !g.PaletteValid[idx] {
				info.MissingPalettePixels++
				continue
			}
			img.SetNRGBA(px, y, g.Palette[idx])
			info.DrawnPixels++
		}
		pos += count
		if count&1 != 0 {
			if pos >= end {
				return nil, info, fmt.Errorf("missing odd-run padding byte in frame %d", frame)
			}
			pos++
		}
		x += count
	}
	return img, info, nil
}

func (g *Graphic) decodeUncompressedFrame(frame, start, end int) (*image.NRGBA, DecodeInfo, error) {
	var info DecodeInfo
	expected := g.Width * g.Height
	if end-start != expected {
		return nil, info, fmt.Errorf("uncompressed frame %d size mismatch: got %d, want %d", frame, end-start, expected)
	}

	img := image.NewNRGBA(image.Rect(0, 0, g.Width, g.Height))
	for i, raw := range g.Data[start:end] {
		idx := int(raw)
		if !g.PaletteValid[idx] {
			info.MissingPalettePixels++
			continue
		}
		x := i % g.Width
		y := i / g.Width
		img.SetNRGBA(x, y, g.Palette[idx])
		info.DrawnPixels++
	}
	return img, info, nil
}
