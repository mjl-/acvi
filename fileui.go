package main

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type fileUI struct {
	column             *columnUI
	file               *os.File // can be nil
	square             *square
	header, body       *duit.Edit
	headerBox, bodyBox *duit.Box
	duit.Box
}

func newFileUI(column *columnUI, filename string) *fileUI {
	slash := strings.HasSuffix(filename, "/")
	if filename != "" {
		filename = path.Clean(filename)
	}
	if slash {
		filename += "/"
	}
	if filename != "" && !strings.HasPrefix(filename, "/") {
		wd, err := os.Getwd()
		if err == nil {
			filename = path.Clean(wd + "/" + filename)
		}
	}
	tag := filename + " Del | "
	taglen := int64(len(tag))
	header, _ := duit.NewEdit(bytes.NewReader([]byte(tag)))
	header.SetCursor(duit.Cursor{Cur: taglen, Start: taglen})
	header.Colors = tagColors
	header.NoScrollbar = true
	height := tagHeight()
	ui := &fileUI{
		column: column,
		header: header,
	}
	ui.square = &square{
		dirty:       false,
		cleanColor:  squareCleanColor,
		borderColor: squareBorderColor,
		dirtyColor:  squareDirtyColor,
		b1: func() {
			ui.column.grow(ui)
		},
		b2: func() {
			ui.column.growAvailable(ui)
		},
		b3: func() {
			ui.column.growFull(ui)
		},
		lowdpiSize: image.Pt(duit.ScrollbarSize, height),
	}
	headerSplit := &duit.Split{
		Split: func(width int) []int {
			w := dui.Scale(duit.ScrollbarSize)
			return []int{w, width - w}
		},
		Kids: duit.NewKids(ui.square, header),
	}
	ui.headerBox = &duit.Box{
		Height: height,
		Width:  -1,
		Kids:   duit.NewKids(headerSplit),
	}
	ui.bodyBox = &duit.Box{
		Width: -1,
		Kids:  duit.NewKids(ui.body),
	}
	ui.init(filename)
	ui.Box.Margin = image.Pt(0, 1)
	ui.Box.Background = squareBorderColor
	ui.Box.Width = -1
	ui.Box.Kids = duit.NewKids(ui.headerBox, ui.bodyBox)

	ui.header.Click = func(m draw.Mouse, offset int64) (e duit.Event) {
		switch m.Buttons {
		case duit.Button1:
		case duit.Button2:
			ui.execute(expandText(ui.header, offset))
		case duit.Button3:
			ui.look(expandText(ui.header, offset))
		}
		return
	}
	ui.header.Keys = func(k rune, m draw.Mouse) (r duit.Event) {
		if k == control&'f' {
			topUI.complete(ui.path(), ui.header)
			r.Consumed = true
		}
		return
	}

	return ui
}

func (ui *fileUI) init(filename string) {
	if filename != "" && !strings.HasSuffix(filename, "/+Errors") {
		fi, err := os.Stat(filename)
		if topUI.error(filename, err, "stat") {
			return
		}
		if fi.IsDir() {
			files, err := ioutil.ReadDir(filename)
			if !topUI.error(filename, err, "readdir") {
				s := ""
				for _, f := range files {
					s += f.Name()
					if f.IsDir() {
						s += "/"
					}
					s += "\n"
				}
				ui.body, _ = duit.NewEdit(bytes.NewReader([]byte(s)))
			}
		} else {
			ui.file, err = os.Open(filename)
			if err == nil {
				ui.body, err = duit.NewEdit(ui.file)
			}
			topUI.error(filename, err, "init")
		}
	}
	if ui.body == nil {
		ui.body, _ = duit.NewEdit(bytes.NewReader([]byte("")))
	}
	ui.body.Colors = textColors
	ui.body.Keys = func(k rune, m draw.Mouse) (r duit.Event) {
		if k == control&'f' {
			topUI.complete(ui.path(), ui.body)
			r.Consumed = true
		}
		return
	}
	ui.body.Click = func(m draw.Mouse, offset int64) (e duit.Event) {
		switch m.Buttons {
		case duit.Button1:
		case duit.Button2:
			ui.execute(expandText(ui.body, offset))
		case duit.Button3:
			ui.look(expandText(ui.body, offset))
		}
		return
	}
	ui.body.DirtyChanged = func(dirty bool) {
		ui.square.dirty = dirty
		dui.MarkDraw(ui.square)
	}
	ui.bodyBox.Kids[0].UI = ui.body
}

func (ui *fileUI) path() string {
	t, _ := ui.header.Text()
	p := strings.Split(string(t), " ")[0]
	// log.Printf("fileUI, path %q, ui.header.Text %q\n", p, ui.header.Text())
	return p
}

func (ui *fileUI) save() {
	p := ui.path()
	if strings.HasSuffix(p, "/") {
		topUI.error(p, fmt.Errorf("is a directory"), "save")
		return
	}
	if strings.HasSuffix(p, "/+Errors") {
		topUI.error(p, fmt.Errorf("is +Errors"), "save")
		return
	}

	// todo: overwrite only the parts that have changed, doing as little buffering as possible
	// todo: get a reader that is independent of the edit state
	buf, err := ioutil.ReadAll(ui.body.Reader())
	if topUI.error(p, err, "read") {
		return
	}
	go func() {
		f, err := os.Create(p)
		if topUI.error(p, err, "create") {
			return
		}
		_, err = f.Write(buf)
		if topUI.error(p, err, "write") {
			err = f.Close()
			topUI.error(p, err, "close")
			return
		}
		err = f.Close()
		topUI.error(p, err, "close")

		dui.Call <- func() {
			ui.body.Saved()
			dui.MarkDraw(ui.body)
		}
	}()
}

func (ui *fileUI) del() {
	ui.column.removeFile(ui)
}

func (ui *fileUI) get() {
	if ui.file != nil {
		ui.file.Close()
		ui.file = nil
	}
	ui.body = nil
	ui.init(ui.path())
	ui.square.dirty = false
	dui.MarkLayout(ui)
}

func (ui *fileUI) execute(t string) {
	switch t {
	case "Put":
		ui.save()
	case "Del":
		ui.del()
	case "Get":
		ui.get()
	default:
		ui.column.execute(ui.path(), t, ui.body)
	}
}

func (ui *fileUI) look(t string) {
	if t == "" {
		return
	}
	if topUI.look(ui.path(), t, true) {
		return
	}

	ui.body.LastSearch = " " + t
	if ui.body.Search(dui, false) {
		ui.body.ScrollCursor(dui)
		dui.MarkDraw(ui.body)
		dui.Call <- func() {
			dui.Focus(ui)
		}
	}
}

func (ui *fileUI) append(buf []byte) {
	ui.body.Append(buf)
	ui.body.ScrollCursor(dui)
	dui.MarkDraw(ui)
}

func (ui *fileUI) Key(dui *duit.DUI, self *duit.Kid, k rune, m draw.Mouse, orig image.Point) (r duit.Result) {
	switch k {
	case draw.KeyCmd + 't':
		dui.Focus(ui.header)
	case draw.KeyCmd + 'm':
		dui.Focus(ui.body)
	case draw.KeyCmd + 's':
		ui.save()
		dui.MarkDraw(ui)
	case draw.KeyCmd + 'w':
		ui.del()
		dui.MarkLayout(ui)
	case draw.KeyCmd + 'e':
		ui.execute(buttonText(ui.header))
	default:
		return ui.Box.Key(dui, self, k, m, orig)
	}
	r.Consumed = true
	return
}

func (ui *fileUI) Focus(dui *duit.DUI, self *duit.Kid, o duit.UI) *image.Point {
	if o == ui {
		p := ui.bodyBox.Focus(dui, ui.Kids[1], ui.body)
		pp := p.Add(ui.Kids[1].R.Min)
		return &pp
	}
	p := ui.headerBox.Focus(dui, ui.Kids[0], o)
	if p == nil {
		p = ui.bodyBox.Focus(dui, ui.Kids[1], o)
		if p != nil {
			pp := p.Add(ui.Kids[1].R.Min)
			return &pp
		}
	}
	return p
}

func (ui *fileUI) Print(self *duit.Kid, indent int) {
	duit.PrintUI("fileUI", self, indent)
	duit.KidsPrint(ui.Kids, indent+1)
}
