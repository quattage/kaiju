package common

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"kaijuengine.com/klib"
)

const BinDir = "../tools/content_tools/"

func MSDFAtlasGenPath() string {
	// Executable names follow msdf-atlas-gen-${OS}-${ARCH}[.exe]
	name := fmt.Sprintf("msdf-atlas-gen-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exePath := filepath.Join(BinDir, name)
	if _, err := os.Stat(exePath); err != nil {
		panic(fmt.Errorf(
			"msdf-atlas-gen not found for %s/%s (expected at %s): %w",
			runtime.GOOS,
			runtime.GOARCH,
			exePath,
			err,
		))
	}
	return exePath
}

func WebfontGenPath() string {
	name := fmt.Sprintf("webfont-ttf-gen-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exePath := filepath.Join(BinDir, name)
	if _, err := os.Stat(exePath); err != nil {
		panic(fmt.Errorf(
			"webfont-generator not found for %s/%s (expected at %s): %w",
			runtime.GOOS,
			runtime.GOARCH,
			exePath,
			err,
		))
	}
	return exePath
}

type Rect struct {
	Left   float32 `json:"left"`
	Top    float32 `json:"top"`
	Right  float32 `json:"right"`
	Bottom float32 `json:"bottom"`
}

type Glyph struct {
	Unicode     int     `json:"unicode"`
	Advance     float32 `json:"advance"`
	PlaneBounds Rect    `json:"planeBounds"`
	AtlasBounds Rect    `json:"atlasBounds"`
}

type Atlas struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
}

type Metrics struct {
	EmSize             float32 `json:"emSize"`
	LineHeight         float32 `json:"lineHeight"`
	Ascender           float32 `json:"ascender"`
	Descender          float32 `json:"descender"`
	UnderlineY         float32 `json:"underlineY"`
	UnderlineThickness float32 `json:"underlineThickness"`
}

type Kerning struct {
	Unicode1 int32   `json:"unicode1"`
	Unicode2 int32   `json:"unicode2"`
	Advance  float32 `json:"advance"`
}

type FontData struct {
	Glyphs  []Glyph   `json:"glyphs"`
	Atlas   Atlas     `json:"atlas"`
	Metrics Metrics   `json:"metrics"`
	Kerning []Kerning `json:"kerning"`
}

func GenerateSDFAtlasFromTTF(ttfName string) {
	println("Processing", ttfName)

	name := filepath.Base(ttfName)
	ttfDir := filepath.Dir(ttfName)
	ttfFile := filepath.Join(ttfDir, name+".ttf")
	jsonFile := filepath.Join(ttfDir, "out", name+".json")
	binFile := filepath.Join(ttfDir, "out", name+".bin")
	pngFile := filepath.Join(ttfDir, "out", name+".png")
	charSetFile := filepath.Join(ttfDir, "charset.txt")

	cmd := exec.Command(MSDFAtlasGenPath(),
		"-font", ttfFile,
		"-pxrange", "4",
		"-size", "64",
		"-charset", charSetFile,
		"-fontname", ttfName,
		"-type", "msdf",
		"-format", "png",
		"-pots",
		"-json", jsonFile,
		"-imageout", pngFile,
	)

	out := klib.MustReturn(cmd.StdoutPipe())
	scanner := bufio.NewScanner(out)
	klib.Must(cmd.Start())
	for scanner.Scan() {
		slog.Info(scanner.Text())
	}
	klib.Must(cmd.Wait())
	jsonBin := klib.MustReturn(os.ReadFile(jsonFile))
	fout := klib.MustReturn(os.Create(binFile))
	defer fout.Close()
	var fontData FontData
	klib.Must(json.NewDecoder(strings.NewReader(string(jsonBin))).Decode(&fontData))
	binary.Write(fout, binary.LittleEndian, int32(len(fontData.Glyphs)))
	binary.Write(fout, binary.LittleEndian, fontData.Atlas.Width)
	binary.Write(fout, binary.LittleEndian, fontData.Atlas.Height)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.EmSize)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.LineHeight)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.Ascender)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.Descender)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.UnderlineY)
	binary.Write(fout, binary.LittleEndian, fontData.Metrics.UnderlineThickness)
	for _, glyph := range fontData.Glyphs {
		binary.Write(fout, binary.LittleEndian, int32(glyph.Unicode))
		binary.Write(fout, binary.LittleEndian, glyph.Advance)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Left)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Top)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Right)
		binary.Write(fout, binary.LittleEndian, glyph.PlaneBounds.Bottom)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Left)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Top)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Right)
		binary.Write(fout, binary.LittleEndian, glyph.AtlasBounds.Bottom)
	}
	os.Remove(jsonFile)
}

type Codepoints struct {
	contents   map[string]rune
	latestFree rune
}

func (m *Codepoints) add(name string) (rune, bool, error) {
	name = strings.TrimSpace(name)
	if code, exists := m.contents[name]; exists {
		return code, false, nil
	}
	if m.latestFree > 0xF8FF {
		return 0, false, fmt.Errorf("icon rune codepoints exhausted")
	}
	code := m.latestFree
	m.contents[name] = code
	m.latestFree++
	return code, true, nil
}

func newManifest(path string, overwrite bool) (*Codepoints, error) {
	path = filepath.Join(path, "codepoints.json")
	m := &Codepoints{
		contents:   make(map[string]rune),
		latestFree: 0xE000,
	}

	if overwrite {
		os.Remove(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := m.save(path); err != nil {
			return nil, err
		}
		return m, nil
	}
	if err := json.Unmarshal(data, &m.contents); err != nil {
		return nil, fmt.Errorf("invalid icon manifest: %w", err)
	}
	for _, code := range m.contents {
		if code >= m.latestFree {
			m.latestFree = code + 1
		}
	}
	return m, nil
}

func (m *Codepoints) save(path string) error {
	data, err := json.MarshalIndent(m.contents, "", "    ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *Codepoints) sync(iconDir string) error {
	entries, err := os.ReadDir(iconDir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	changed := false
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".svg" {
			continue
		}
		name := entry.Name()[:len(entry.Name())-4]
		_, added, err := m.add(name)
		if err != nil {
			return err
		}
		changed = changed || added
	}
	if changed {
		return m.save(filepath.Join(iconDir, "codepoints.json"))
	}
	return nil
}

func (m *Codepoints) MakeCharset(iconDir string) error {
	iconDir = filepath.Join(iconDir, "charset.txt")
	codes := make([]int, 0, len(m.contents))
	for _, code := range m.contents {
		codes = append(codes, int(code))
	}
	sort.Ints(codes)
	var out strings.Builder
	for _, code := range codes {
		fmt.Fprintf(&out, "0x%X\n", code)
	}
	return os.WriteFile(iconDir, []byte(out.String()), 0644)
}

func GenerateTTFFromSVGs(directory string, fontName string, writeCharset bool) error {
	manif, err := newManifest(directory, true)
	if err != nil {
		return err
	}
	err = manif.sync(directory)
	if err != nil {
		return err
	}
	cmd := exec.Command(WebfontGenPath(),
		"--dest", "../icons",
		"--font-name", fontName,
		"--types", "ttf",
		"--no-css",
		"--start-codepoint", "0xE000",
		directory,
	)
	out := klib.MustReturn(cmd.StdoutPipe())
	scanner := bufio.NewScanner(out)
	klib.Must(cmd.Start())
	for scanner.Scan() {
		slog.Info(scanner.Text())
	}
	klib.Must(cmd.Wait())
	if writeCharset {
		slog.Info(fmt.Sprintf("Writing charset for %d icons...", len(manif.contents)))
		err := manif.MakeCharset(directory)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateSDFAtlasFromSVGs(directory string) error {
	if err := GenerateTTFFromSVGs(directory, "TempIcons", true); err != nil {
		return errors.Join(fmt.Errorf("Failed while generating TTF - "), err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "out"), 0755); err != nil {
		return err
	}
	GenerateSDFAtlasFromTTF(filepath.Join(directory, "TempIcons"))
	os.Remove(filepath.Join(directory, "TempIcons.ttf"))
	os.Remove(filepath.Join(directory, "charset.txt"))
	return nil
}
