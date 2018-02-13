package main

import (
	"bytes"
	"image"
	"log"
	"os"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type columnUI struct {
	header    *duit.Edit
	headerBox *duit.Box
	files     *filesUI
	duit.Box
}

func newColumnUI(paths []string) *columnUI {
	tag := "New Delcol "
	header, _ := duit.NewEdit(bytes.NewReader([]byte(tag)))
	taglen := int64(len(tag))
	header.SetCursor(duit.Cursor{Cur: taglen, Start: taglen})
	header.Colors = tagColors
	header.NoScrollbar = true
	ui := &columnUI{
		header: header,
	}
	header.Click = func(m draw.Mouse, offset int64) (e duit.Event) {
		switch m.Buttons {
		case duit.Button1:
		case duit.Button2:
			ui.execute("", expandText(header, offset), nil)
		case duit.Button3:
			ui.look("", expandText(header, offset))
		}
		return
	}
	header.Keys = func(k rune, m draw.Mouse) (r duit.Event) {
		if k == control&'f' {
			p, _ := os.Getwd()
			topUI.complete(p, ui.header)
			r.Consumed = true
		}
		return
	}
	square := &square{
		dirty:       false,
		cleanColor:  squareBorderColor,
		borderColor: squareBorderColor,
		dirtyColor:  squareBorderColor,
		b1: func() {
			topUI.grow(ui)
		},
		b2: func() {
			topUI.growAvailable(ui)
		},
		b3: func() {
			topUI.growFull(ui)
		},
		lowdpiSize: image.Pt(duit.ScrollbarSize, tagHeight()),
	}
	headerSplit := &duit.Split{
		Split: func(width int) []int {
			w := dui.Scale(duit.ScrollbarSize)
			return []int{w, width - w}
		},
		Kids: duit.NewKids(square, header),
	}
	ui.headerBox = &duit.Box{
		Width:  -1,
		Height: tagHeight(),
		Kids:   duit.NewKids(headerSplit),
	}
	var files []*fileUI
	for _, p := range paths {
		files = append(files, newFileUI(ui, p))
	}
	ui.files = newFilesUI(ui, files)
	ui.Box.Margin = image.Pt(0, 1)
	ui.Box.Background = textColors.Fg
	ui.Box.Width = -1
	ui.Box.Height = -1
	ui.Box.Kids = duit.NewKids(ui.headerBox, ui.files)
	return ui
}

func (ui *columnUI) execute(filename, t string, edit *duit.Edit) {
	switch t {
	case "New":
		ui.addFile("")
	case "Delcol":
		topUI.removeColumn(ui)
		dui.MarkLayout(topUI)
	default:
		topUI.execute(filename, t, edit)
	}
}

func (ui *columnUI) look(filename, t string) {
	log.Printf("columnUI.look %q\n", t)
	topUI.look(filename, t, true)
}

func (ui *columnUI) addFile(filename string) *fileUI {
	f := newFileUI(ui, filename)
	ui.files.add(f)
	return f
}

func (ui *columnUI) removeFile(file *fileUI) {
	ui.files.remove(file)
}

func (ui *columnUI) fileIndexMouse(m draw.Mouse) int {
	y := ui.Kids[1].R.Min.Y
	marginY := ui.files.marginY()
	for i, h := range ui.files.heights {
		if m.Y >= y && m.Y < y+h {
			return i
		}
		y += h + marginY
	}
	return -1
}

func (ui *columnUI) fileIndex(file *fileUI) int {
	for i, f := range ui.files.files {
		if f == file {
			return i
		}
	}
	return -1
}

func (ui *columnUI) grow(file *fileUI) {
	i := ui.fileIndex(file)
	if i >= 0 {
		ui.growIndex(i)
	}
}

func (ui *columnUI) growAvailable(file *fileUI) {
	i := ui.fileIndex(file)
	if i >= 0 {
		take := (ui.files.height() - len(ui.files.files)*tagHeight()) - ui.files.heights[i]
		height := tagHeight()
		grow(take, i, ui.files.heights, height)
		setMinDims(ui.files.heights, height)
		dui.MarkLayout(ui.files)
	}
}

func (ui *columnUI) growFull(file *fileUI) {
	i := ui.fileIndex(file)
	if i >= 0 {
		take := ui.files.height() - ui.files.heights[i]
		grow(take, i, ui.files.heights, tagHeight())
		dui.MarkLayout(ui.files)
	}
}

func (ui *columnUI) growIndex(index int) {
	height := tagHeight()
	grow(dui.Scale(100), index, ui.files.heights, height)
	setMinDims(ui.files.heights, height)
	dui.MarkLayout(ui.files)
}

func (ui *columnUI) Key(dui *duit.DUI, self *duit.Kid, k rune, m draw.Mouse, orig image.Point) (r duit.Result) {
	switch k {
	case draw.KeyCmd + 'I':
		i := ui.fileIndexMouse(m)
		if i < 0 {
			return
		}
		ui.growIndex(i)
	case draw.KeyCmd + 'J':
		i := (ui.fileIndexMouse(m) + 1) % len(ui.files.heights)
		dui.Focus(ui.files.files[i])
	case draw.KeyCmd + 'K':
		i := (ui.fileIndexMouse(m) - 1 + len(ui.files.heights)) % len(ui.files.heights)
		dui.Focus(ui.files.files[i])
	case draw.KeyCmd + 'n':
		ui.addFile("")
	default:
		return ui.Box.Key(dui, self, k, m, orig)
	}
	r.Consumed = true
	return
}

func (ui *columnUI) Print(self *duit.Kid, indent int) {
	duit.PrintUI("columnUI", self, indent)
	duit.KidsPrint(ui.Kids, indent+1)
}
