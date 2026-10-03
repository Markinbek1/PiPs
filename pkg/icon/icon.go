package icon

import (
	"bytes"
	"encoding/binary"
)

// GenerateICO builds a 16x16 32-bit RGBA ICO file in memory.
// r, g, b specify the main indicator color.
func GenerateICO(r, g, b byte, transparentCenter bool) []byte {
	var buf bytes.Buffer

	// 1. ICO Header (6 bytes)
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Type = Icon
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Image count = 1

	// 2. Directory Entry (16 bytes)
	imageSize := uint32(40 + 1024 + 32) // BMP Header (40) + RGBA Pixels (1024) + AND Mask (32)
	offset := uint32(22)

	buf.WriteByte(16)                                  // Width 16
	buf.WriteByte(16)                                  // Height 16
	buf.WriteByte(0)                                   // Colors
	buf.WriteByte(0)                                   // Reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // Color planes
	binary.Write(&buf, binary.LittleEndian, uint16(32))// Bits per pixel
	binary.Write(&buf, binary.LittleEndian, imageSize) // Size of image data
	binary.Write(&buf, binary.LittleEndian, offset)    // Offset of BMP data

	// 3. BITMAPINFOHEADER (40 bytes)
	binary.Write(&buf, binary.LittleEndian, uint32(40))   // Header size
	binary.Write(&buf, binary.LittleEndian, int32(16))    // Width
	binary.Write(&buf, binary.LittleEndian, int32(32))    // Height (X2 for mask)
	binary.Write(&buf, binary.LittleEndian, uint16(1))    // Planes
	binary.Write(&buf, binary.LittleEndian, uint16(32))   // BitCount
	binary.Write(&buf, binary.LittleEndian, uint32(0))    // Compression (BI_RGB)
	binary.Write(&buf, binary.LittleEndian, uint32(1024)) // Image size
	binary.Write(&buf, binary.LittleEndian, int32(0))     // XPels
	binary.Write(&buf, binary.LittleEndian, int32(0))     // YPels
	binary.Write(&buf, binary.LittleEndian, uint32(0))    // Colors used
	binary.Write(&buf, binary.LittleEndian, uint32(0))    // Important colors

	// 4. RGBA Pixels (16x16, bottom-up ordering)
	pixels := make([]byte, 1024)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			// Center distance calculation for circle
			dx := float64(x) - 7.5
			dy := float64(y) - 7.5
			distSq := dx*dx + dy*dy

			idx := ((15 - y) * 16 + x) * 4 // BGRA format in BMP

			if distSq <= 36.0 {
				if transparentCenter && distSq < 16.0 {
					// Hollow center for click-through mode
					pixels[idx+0] = 0 // B
					pixels[idx+1] = 0 // G
					pixels[idx+2] = 0 // R
					pixels[idx+3] = 0 // Alpha transparent
				} else {
					pixels[idx+0] = b
					pixels[idx+1] = g
					pixels[idx+2] = r
					pixels[idx+3] = 255 // Alpha opaque
				}
			} else {
				pixels[idx+0] = 0
				pixels[idx+1] = 0
				pixels[idx+2] = 0
				pixels[idx+3] = 0
			}
		}
	}
	buf.Write(pixels)

	// 5. AND Mask (32 bytes of 0s)
	andMask := make([]byte, 32)
	buf.Write(andMask)

	return buf.Bytes()
}

func GetIconOff() []byte {
	// Sleek Blue/Purple Icon for normal mode (solid circle)
	return GenerateICO(41, 128, 255, false)
}

func GetIconOn() []byte {
	// Vibrant Green/Emerald Icon with hollow center for click-through ON
	return GenerateICO(0, 230, 118, true)
}
