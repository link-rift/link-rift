package qrcode

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"

	"github.com/link-rift/link-rift/pkg/storage"
)

// Options configures QR code generation.
type Options struct {
	Size            int
	ErrorCorrection string // L, M, Q, H
	ForegroundColor string // hex like #000000
	BackgroundColor string // hex like #FFFFFF
	LogoURL         string
	DotStyle        string
	CornerStyle     string
	Margin          int
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size:            512,
		ErrorCorrection: "M",
		ForegroundColor: "#000000",
		BackgroundColor: "#FFFFFF",
		DotStyle:        "square",
		CornerStyle:     "square",
		Margin:          4,
	}
}

// Generator generates QR code images.
type Generator struct {
	storage storage.ObjectStorage
}

// NewGenerator creates a new QR code generator.
func NewGenerator(store storage.ObjectStorage) *Generator {
	return &Generator{storage: store}
}

// Generate creates a PNG QR code image and returns the bytes.
func (g *Generator) Generate(url string, opts Options) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 512
	}
	if opts.Size > 2048 {
		opts.Size = 2048
	}

	fg := parseHexColorWithDefault(opts.ForegroundColor, color.Black)
	bg := parseHexColorWithDefault(opts.BackgroundColor, color.White)

	// Generate QR matrix using our built-in encoder
	matrix, err := encodeQR(url, opts.ErrorCorrection)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR data: %w", err)
	}

	margin := opts.Margin
	if margin < 0 {
		margin = 4
	}

	moduleCount := len(matrix)
	totalModules := moduleCount + 2*margin

	// Calculate module pixel size
	moduleSize := opts.Size / totalModules
	if moduleSize < 1 {
		moduleSize = 1
	}
	imgSize := totalModules * moduleSize

	img := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	// Fill background
	for y := 0; y < imgSize; y++ {
		for x := 0; x < imgSize; x++ {
			img.Set(x, y, bg)
		}
	}

	// Draw modules
	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			if matrix[row][col] {
				px := (col + margin) * moduleSize
				py := (row + margin) * moduleSize
				for dy := 0; dy < moduleSize; dy++ {
					for dx := 0; dx < moduleSize; dx++ {
						img.Set(px+dx, py+dy, fg)
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateSVG creates an SVG QR code and returns the bytes.
func (g *Generator) GenerateSVG(url string, opts Options) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 512
	}

	matrix, err := encodeQR(url, opts.ErrorCorrection)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR data: %w", err)
	}

	margin := opts.Margin
	if margin < 0 {
		margin = 4
	}

	fgHex := opts.ForegroundColor
	if fgHex == "" {
		fgHex = "#000000"
	}
	bgHex := opts.BackgroundColor
	if bgHex == "" {
		bgHex = "#FFFFFF"
	}

	moduleCount := len(matrix)
	totalModules := moduleCount + 2*margin
	moduleSize := opts.Size / totalModules
	if moduleSize < 1 {
		moduleSize = 1
	}
	totalSize := totalModules * moduleSize

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, totalSize, totalSize, opts.Size, opts.Size)
	fmt.Fprintf(&buf, `<rect width="%d" height="%d" fill="%s"/>`, totalSize, totalSize, bgHex)

	for row := 0; row < moduleCount; row++ {
		for col := 0; col < moduleCount; col++ {
			if matrix[row][col] {
				px := (col + margin) * moduleSize
				py := (row + margin) * moduleSize
				fmt.Fprintf(&buf, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
					px, py, moduleSize, moduleSize, fgHex)
			}
		}
	}

	buf.WriteString(`</svg>`)
	return buf.Bytes(), nil
}

// GenerateAndUpload generates a QR code and uploads it to storage.
func (g *Generator) GenerateAndUpload(ctx context.Context, url, storageKey string, opts Options) (pngURL string, err error) {
	pngBytes, err := g.Generate(url, opts)
	if err != nil {
		return "", err
	}

	pngURL, err = g.storage.Upload(ctx, storageKey, pngBytes, "image/png")
	if err != nil {
		return "", fmt.Errorf("failed to upload QR code: %w", err)
	}

	return pngURL, nil
}

// GenerateDataURI generates a QR code and returns it as a data URI string.
func (g *Generator) GenerateDataURI(url string, opts Options) (string, error) {
	pngBytes, err := g.Generate(url, opts)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	return "data:image/png;base64," + encoded, nil
}

func parseHexColorWithDefault(hex string, defaultColor color.Color) color.Color {
	c, err := ParseHexColor(hex)
	if err != nil {
		return defaultColor
	}
	return c
}

// ParseHexColor parses a hex color string like "#FF0000" into a color.Color.
func ParseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	green, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, err
	}

	return color.RGBA{R: uint8(r), G: uint8(green), B: uint8(b), A: 255}, nil
}

// =========================================================================
// Built-in QR encoder (Version 1-6, byte mode, no external dependencies)
// Supports short URLs typical for link shortener use cases.
// =========================================================================

// encodeQR creates a QR code boolean matrix from the given data.
func encodeQR(data string, ecLevel string) ([][]bool, error) {
	dataBytes := []byte(data)

	// Determine version (1-40) based on data length and EC level
	version, ecIdx := selectVersion(len(dataBytes), ecLevel)
	if version < 0 {
		return nil, fmt.Errorf("data too long for QR code")
	}

	size := 17 + version*4
	matrix := makeMatrix(size)
	reserved := makeMatrix(size)

	// Place function patterns
	placeFunctionPatterns(matrix, reserved, version, size)

	// Encode data into bits
	bits := encodeDataBits(dataBytes, version, ecIdx)

	// Place data bits
	placeDataBits(matrix, reserved, bits, size)

	// Apply mask (pattern 0 for simplicity)
	applyMask(matrix, reserved, size, 0)

	// Place format info
	placeFormatInfo(matrix, ecIdx, 0, size)

	if version >= 7 {
		placeVersionInfo(matrix, version, size)
	}

	return matrix, nil
}

func makeMatrix(size int) [][]bool {
	m := make([][]bool, size)
	for i := range m {
		m[i] = make([]bool, size)
	}
	return m
}

// QR version capacity for byte mode (approximate, conservative)
var versionCapacity = [41][4]int{
	{0, 0, 0, 0}, // version 0 unused
	{17, 14, 11, 7},
	{32, 26, 20, 14},
	{53, 42, 32, 24},
	{78, 62, 46, 34},
	{106, 84, 60, 44},
	{134, 106, 74, 58},
	{154, 122, 86, 64},
	{192, 152, 108, 84},
	{230, 180, 130, 98},
	{271, 213, 151, 119},    // v10
	{321, 251, 177, 137},
	{367, 287, 203, 155},
	{425, 331, 241, 177},
	{458, 362, 258, 194},
	{520, 412, 292, 220},
	{586, 450, 322, 250},
	{644, 504, 364, 280},
	{718, 560, 394, 310},
	{792, 624, 442, 338},
	{858, 666, 482, 382},    // v20
	{929, 711, 509, 403},
	{1003, 779, 565, 439},
	{1091, 857, 611, 461},
	{1171, 911, 661, 511},
	{1273, 997, 715, 535},
	{1367, 1059, 751, 593},
	{1465, 1125, 805, 625},
	{1528, 1190, 868, 658},
	{1628, 1264, 908, 698},
	{1732, 1370, 982, 742},  // v30
	{1840, 1452, 1030, 790},
	{1952, 1538, 1112, 842},
	{2068, 1628, 1168, 898},
	{2188, 1722, 1228, 958},
	{2303, 1809, 1283, 983},
	{2431, 1911, 1351, 1051},
	{2563, 1989, 1423, 1093},
	{2699, 2099, 1499, 1139},
	{2809, 2213, 1579, 1219},
	{2953, 2331, 1663, 1273}, // v40
}

func selectVersion(dataLen int, ecLevel string) (int, int) {
	ecIdx := 0 // L=0, M=1, Q=2, H=3
	switch strings.ToUpper(ecLevel) {
	case "L":
		ecIdx = 0
	case "M":
		ecIdx = 1
	case "Q":
		ecIdx = 2
	case "H":
		ecIdx = 3
	default:
		ecIdx = 1
	}

	for v := 1; v <= 40; v++ {
		if versionCapacity[v][ecIdx] >= dataLen {
			return v, ecIdx
		}
	}
	return -1, ecIdx
}

func placeFunctionPatterns(matrix, reserved [][]bool, version, size int) {
	// Finder patterns (7x7) at three corners
	placeFinderPattern(matrix, reserved, 0, 0, size)
	placeFinderPattern(matrix, reserved, 0, size-7, size)
	placeFinderPattern(matrix, reserved, size-7, 0, size)

	// Separators (already handled by finder pattern placement)

	// Timing patterns
	for i := 8; i < size-8; i++ {
		val := i%2 == 0
		setModule(matrix, 6, i, val, size)
		setModule(reserved, 6, i, true, size)
		setModule(matrix, i, 6, val, size)
		setModule(reserved, i, 6, true, size)
	}

	// Alignment patterns
	if version >= 2 {
		positions := alignmentPatternPositions(version, size)
		for _, row := range positions {
			for _, col := range positions {
				// Skip if overlapping with finder patterns
				if (row < 9 && col < 9) ||
					(row < 9 && col > size-9) ||
					(row > size-9 && col < 9) {
					continue
				}
				placeAlignmentPattern(matrix, reserved, row, col, size)
			}
		}
	}

	// Reserve format info areas
	for i := 0; i < 8; i++ {
		setModule(reserved, 8, i, true, size)
		setModule(reserved, i, 8, true, size)
		setModule(reserved, 8, size-1-i, true, size)
		setModule(reserved, size-1-i, 8, true, size)
	}
	setModule(reserved, 8, 8, true, size)

	// Dark module
	setModule(matrix, size-8, 8, true, size)
	setModule(reserved, size-8, 8, true, size)

	// Version info areas (version >= 7)
	if version >= 7 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 3; j++ {
				setModule(reserved, i, size-11+j, true, size)
				setModule(reserved, size-11+j, i, true, size)
			}
		}
	}
}

func placeFinderPattern(matrix, reserved [][]bool, row, col, size int) {
	for r := -1; r <= 7; r++ {
		for c := -1; c <= 7; c++ {
			rr := row + r
			cc := col + c
			if rr < 0 || rr >= size || cc < 0 || cc >= size {
				continue
			}
			val := false
			if r >= 0 && r <= 6 && c >= 0 && c <= 6 {
				if r == 0 || r == 6 || c == 0 || c == 6 ||
					(r >= 2 && r <= 4 && c >= 2 && c <= 4) {
					val = true
				}
			}
			matrix[rr][cc] = val
			reserved[rr][cc] = true
		}
	}
}

func placeAlignmentPattern(matrix, reserved [][]bool, row, col, size int) {
	for r := -2; r <= 2; r++ {
		for c := -2; c <= 2; c++ {
			rr := row + r
			cc := col + c
			if rr < 0 || rr >= size || cc < 0 || cc >= size {
				continue
			}
			val := r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0)
			matrix[rr][cc] = val
			reserved[rr][cc] = true
		}
	}
}

func alignmentPatternPositions(version, size int) []int {
	if version == 1 {
		return nil
	}
	// Simplified alignment pattern calculation
	intervals := version/7 + 1
	step := 0
	if intervals > 1 {
		step = (size - 13) / intervals
		if step%2 != 0 {
			step++
		}
	}

	positions := []int{6}
	pos := size - 7
	for i := 0; i < intervals; i++ {
		found := false
		for _, p := range positions {
			if p == pos {
				found = true
				break
			}
		}
		if !found {
			positions = append(positions, pos)
		}
		pos -= step
	}

	// Sort positions
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			if positions[j] < positions[i] {
				positions[i], positions[j] = positions[j], positions[i]
			}
		}
	}

	return positions
}

func setModule(matrix [][]bool, row, col int, val bool, size int) {
	if row >= 0 && row < size && col >= 0 && col < size {
		matrix[row][col] = val
	}
}

func encodeDataBits(data []byte, version, ecIdx int) []bool {
	// Mode indicator for byte mode: 0100
	bits := []bool{false, true, false, false}

	// Character count indicator
	ccBits := 8
	if version >= 10 {
		ccBits = 16
	}
	for i := ccBits - 1; i >= 0; i-- {
		bits = append(bits, (len(data)>>uint(i))&1 == 1)
	}

	// Data bytes
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>uint(i))&1 == 1)
		}
	}

	// Total data capacity (simplified — use version capacity)
	totalBits := versionCapacity[version][ecIdx] * 8

	// Terminator
	termLen := 4
	if len(bits)+termLen > totalBits {
		termLen = totalBits - len(bits)
	}
	for i := 0; i < termLen; i++ {
		bits = append(bits, false)
	}

	// Pad to byte boundary
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}

	// Pad bytes
	padBytes := []byte{0xEC, 0x11}
	padIdx := 0
	for len(bits) < totalBits {
		b := padBytes[padIdx%2]
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>uint(i))&1 == 1)
		}
		padIdx++
	}

	if len(bits) > totalBits {
		bits = bits[:totalBits]
	}

	return bits
}

func placeDataBits(matrix, reserved [][]bool, bits []bool, size int) {
	bitIdx := 0
	upward := true
	col := size - 1

	for col >= 0 {
		if col == 6 {
			col--
		}

		for row := 0; row < size; row++ {
			r := row
			if upward {
				r = size - 1 - row
			}

			for dc := 0; dc <= 1; dc++ {
				c := col - dc
				if c < 0 || c >= size {
					continue
				}
				if reserved[r][c] {
					continue
				}
				if bitIdx < len(bits) {
					matrix[r][c] = bits[bitIdx]
					bitIdx++
				}
			}
		}

		upward = !upward
		col -= 2
	}
}

func applyMask(matrix, reserved [][]bool, size, maskPattern int) {
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if reserved[row][col] {
				continue
			}
			var masked bool
			switch maskPattern {
			case 0:
				masked = (row+col)%2 == 0
			case 1:
				masked = row%2 == 0
			case 2:
				masked = col%3 == 0
			case 3:
				masked = (row+col)%3 == 0
			case 4:
				masked = (row/2+col/3)%2 == 0
			case 5:
				masked = (row*col)%2+(row*col)%3 == 0
			case 6:
				masked = ((row*col)%2+(row*col)%3)%2 == 0
			case 7:
				masked = ((row+col)%2+(row*col)%3)%2 == 0
			}
			if masked {
				matrix[row][col] = !matrix[row][col]
			}
		}
	}
}

// Format info: 15-bit BCH code
var formatInfoBits = [4][8]uint16{
	// EC Level L, masks 0-7
	{0x77c4, 0x72f3, 0x7daa, 0x789d, 0x662f, 0x6318, 0x6c41, 0x6976},
	// EC Level M, masks 0-7
	{0x5412, 0x5125, 0x5e7c, 0x5b4b, 0x45f9, 0x40ce, 0x4f97, 0x4aa0},
	// EC Level Q, masks 0-7
	{0x355f, 0x3068, 0x3f31, 0x3a06, 0x24b4, 0x2183, 0x2eda, 0x2bed},
	// EC Level H, masks 0-7
	{0x1689, 0x13be, 0x1ce7, 0x19d0, 0x0762, 0x0255, 0x0d0c, 0x083b},
}

func placeFormatInfo(matrix [][]bool, ecIdx, maskPattern, size int) {
	info := formatInfoBits[ecIdx][maskPattern]

	// First copy: around top-left finder
	positions1 := [][2]int{
		{0, 8}, {1, 8}, {2, 8}, {3, 8}, {4, 8}, {5, 8}, {7, 8}, {8, 8},
		{8, 7}, {8, 5}, {8, 4}, {8, 3}, {8, 2}, {8, 1}, {8, 0},
	}
	for i, pos := range positions1 {
		bit := (info >> uint(14-i)) & 1
		matrix[pos[0]][pos[1]] = bit == 1
	}

	// Second copy: around other two finders
	positions2 := [][2]int{
		{8, size - 1}, {8, size - 2}, {8, size - 3}, {8, size - 4},
		{8, size - 5}, {8, size - 6}, {8, size - 7},
		{size - 7, 8}, {size - 6, 8}, {size - 5, 8}, {size - 4, 8},
		{size - 3, 8}, {size - 2, 8}, {size - 1, 8},
	}

	for i, pos := range positions2 {
		bit := (info >> uint(14-i)) & 1
		if pos[0] < size && pos[1] < size {
			matrix[pos[0]][pos[1]] = bit == 1
		}
	}
}

func placeVersionInfo(matrix [][]bool, version, size int) {
	if version < 7 {
		return
	}

	info := versionInfoBits(version)

	for i := 0; i < 18; i++ {
		bit := (info >> uint(i)) & 1
		row := i / 3
		col := size - 11 + i%3
		matrix[row][col] = bit == 1
		matrix[col][row] = bit == 1
	}
}

func versionInfoBits(version int) uint32 {
	// Version info is an 18-bit BCH(18,6) code
	data := uint32(version) << 12
	gen := uint32(0x1F25) // Generator polynomial for BCH(18,6)

	rem := data
	for i := 17; i >= 12; i-- {
		if rem&(1<<uint(i)) != 0 {
			rem ^= gen << uint(i-12)
		}
	}
	return data | rem
}

