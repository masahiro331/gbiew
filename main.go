package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

type FileInfo struct {
	f      *os.File
	offset int64
}

const (
	coldef       = termbox.ColorDefault
	LineSize     = 69
	LineByteSize = 16
)

func drawFullView(f *FileInfo, offset int) {
	_, h := termbox.Size()
	h = h - 2
	viewSize := h * LineByteSize

	if offset == 0 {
		f.f.Seek(0, 0)
		f.offset = 0
	}

	if offset == -1 {
		fi, err := f.f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		f.f.Seek(-int64(viewSize), 2)
		f.offset = fi.Size() - int64(viewSize)
	}

	buf := make([]byte, viewSize)
	_, err := f.f.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	termbox.Clear(coldef, coldef)
	for y := 0; y < h; y++ {
		for x, r := range lineStr(buf[LineByteSize*y:LineByteSize*(y+1)], uint64(f.offset)) {
			termbox.SetCell(x, y, r, coldef, coldef)
		}
		f.offset = f.offset + LineByteSize
	}
	termbox.Flush()
}

func drawResetView(f *FileInfo) {
	_, h := termbox.Size()
	h = h - 2
	viewSize := h * LineByteSize

	buf := make([]byte, viewSize)
	_, err := f.f.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	termbox.Clear(coldef, coldef)
	for y := 0; y < h; y++ {
		f.offset = f.offset + LineByteSize
		for x, r := range lineStr(buf[LineByteSize*y:LineByteSize*(y+1)], uint64(f.offset)) {
			termbox.SetCell(x, y, r, coldef, coldef)
		}
	}
	termbox.Flush()
}

func drawDownView(f *FileInfo, offset int, forward bool) error {
	_, h := termbox.Size()
	h = h - 2
	viewSize := h * LineByteSize

	if forward {
		offset = viewSize
	}
	_, err := f.f.Seek(int64(-viewSize+offset), 1)
	if err != nil {
		return err
	}

	buf := make([]byte, viewSize)

	_, err = f.f.Read(buf)
	if err != nil {
		return err
	}

	f.offset = f.offset - int64(viewSize-offset)
	termbox.Clear(coldef, coldef)
	for y := 0; y < h; y++ {
		f.offset = f.offset + LineByteSize
		for x, r := range lineStr(buf[LineByteSize*y:LineByteSize*(y+1)], uint64(f.offset)) {
			termbox.SetCell(x, y, r, coldef, coldef)
		}
	}
	termbox.Flush()
	return nil
}

func drawUpView(f *FileInfo, offset int, back bool) error {
	_, h := termbox.Size()
	h = h - 2
	viewSize := h * LineByteSize

	if back {
		offset = viewSize
	}
	_, err := f.f.Seek(int64(-viewSize-offset), 1)
	if err != nil {
		return err
	}

	buf := make([]byte, viewSize)
	_, err = f.f.Read(buf)
	if err != nil {
		return err
	}

	f.offset = f.offset - int64(viewSize+offset)
	termbox.Clear(coldef, coldef)
	for y := 0; y < h; y++ {
		f.offset = f.offset + LineByteSize
		for x, r := range lineStr(buf[LineByteSize*y:LineByteSize*(y+1)], uint64(f.offset)) {
			termbox.SetCell(x, y, r, coldef, coldef)
		}
	}
	termbox.Flush()
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("invalid arguments")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	mainLoop(&FileInfo{
		f:      f,
		offset: 0,
	})
}

func mainLoop(f *FileInfo) {
	var put_g bool

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	drawFullView(f, 0)
MAINLOOP:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break MAINLOOP
			case termbox.KeyCtrlF:
				err := drawDownView(f, LineByteSize, true)
				if err != nil {
					break
				}
			case termbox.KeyCtrlB:
				err := drawUpView(f, LineByteSize, true)
				if err != nil {
					break
				}
			}

			switch ev.Ch {
			case 'j':
				err := drawDownView(f, LineByteSize, false)
				if err != nil {
					break
				}
			case 'k':
				err := drawUpView(f, LineByteSize, false)
				if err != nil {
					break
				}
			case 'g':
				if put_g {
					drawFullView(f, 0)
					put_g = false
				} else {
					put_g = true
				}
			case 'G':
				drawFullView(f, -1)

			// Command
			case ':':

			// Search
			case '/':
				drawFullView(f, -1)
			}
		}
	}
}

func drawCommand(f *FileInfo, offset int) {
}

func drawSearch(f *FileInfo, offset int) {
}

func lineStr(buf []byte, offset uint64) []rune {
	var binaryStr string
	for i, b := range buf {
		if i%2 == 0 {
			binaryStr = binaryStr + " "
		}
		binaryStr = binaryStr + fmt.Sprintf("%02x", b)
	}
	return []rune(fmt.Sprintf("%08x: %s  %s", offset, binaryStr, formatBinaryString(buf)))
}

func formatBinaryString(buf []byte) (str string) {
	for _, b := range buf {
		if b > 0x20 && b < 0x7f {
			str = str + string(b)
		} else {
			str = str + "."
		}
	}
	return
}
