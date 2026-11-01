// Command makeico builds icon/WebIcon.ico from icon/WebIcon.png.
//
// The .ico holds PNG-compressed entries at several sizes, so Windows can pick
// the right one for the tray, the taskbar and Explorer. Run it from the repo
// root after changing the artwork:
//
//	go run ./icon/makeico
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/draw"
)

const (
	srcPath = "icon/WebIcon.png"
	dstPath = "icon/WebIcon.ico"
)

var sizes = []int{16, 24, 32, 48, 64, 128, 256}

func main() {
	f, err := os.Open(srcPath)
	check(err)
	defer f.Close()
	src, err := png.Decode(f)
	check(err)

	var images [][]byte
	for _, s := range sizes {
		dst := image.NewNRGBA(image.Rect(0, 0, s, s))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
		var buf bytes.Buffer
		check(png.Encode(&buf, dst))
		images = append(images, buf.Bytes())
	}

	// ICONDIR header, then one ICONDIRENTRY per image, then the image data.
	var out []byte
	out = binary.LittleEndian.AppendUint16(out, 0) // reserved
	out = binary.LittleEndian.AppendUint16(out, 1) // type: icon
	out = binary.LittleEndian.AppendUint16(out, uint16(len(images)))

	offset := uint32(6 + 16*len(images))
	for i, data := range images {
		size := sizes[i]
		dim := byte(size)
		if size >= 256 {
			dim = 0 // 0 means 256 in the ICO format
		}
		out = append(out, dim, dim, 0, 0)               // width, height, palette size, reserved
		out = binary.LittleEndian.AppendUint16(out, 1)  // colour planes
		out = binary.LittleEndian.AppendUint16(out, 32) // bits per pixel
		out = binary.LittleEndian.AppendUint32(out, uint32(len(data)))
		out = binary.LittleEndian.AppendUint32(out, offset)
		offset += uint32(len(data))
	}
	for _, data := range images {
		out = append(out, data...)
	}

	check(os.WriteFile(dstPath, out, 0o644))
	log.Printf("wrote %s (%d sizes, %d bytes)", dstPath, len(sizes), len(out))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
