package main

import (
	"fmt"
	"image"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type square struct {
	dirty                               bool
	cleanColor, borderColor, dirtyColor *draw.Image
	b1, b2, b3                          func()
	lowdpiSize, size                    image.Point
	m                                   draw.Mouse
}

var _ duit.UI = &square{}

func (ui *square) Layout(dui *duit.DUI, self *duit.Kid, sizeAvail image.Point, force bool) {
	x := dui.Scale(ui.lowdpiSize.X)
	y := dui.Scale(ui.lowdpiSize.Y)
	ui.size = image.Pt(x, y)
	self.Layout = duit.Clean
	self.Draw = duit.Dirty
	self.R = image.Rect(0, 0, x, y)
}

func (ui *square) Draw(dui *duit.DUI, self *duit.Kid, img *draw.Image, orig image.Point, m draw.Mouse, force bool) {
	var bg *draw.Image
	if ui.dirty {
		bg = ui.dirtyColor
	} else {
		bg = ui.cleanColor
	}
	img.Draw(self.R.Add(orig), bg, nil, image.ZP)
	img.Border(self.R.Add(orig), dui.Scale(1), ui.borderColor, image.ZP)
}

func (ui *square) Mouse(dui *duit.DUI, self *duit.Kid, m draw.Mouse, origM draw.Mouse, orig image.Point) (r duit.Result) {
	r.Hit = ui
	om := ui.m
	ui.m = m
	if om.Buttons != 0 && m.Buttons == 0 {
		switch om.Buttons {
		case duit.Button1:
			ui.b1()
		case duit.Button2:
			ui.b2()
		case duit.Button3:
			ui.b3()
		default:
			return
		}
		dui.Call <- func() {
			dui.Focus(ui)
		}
		r.Consumed = true
	}
	return
}

func (ui *square) Key(dui *duit.DUI, self *duit.Kid, k rune, m draw.Mouse, orig image.Point) (r duit.Result) {
	r.Hit = ui
	return
}

func (ui *square) FirstFocus(dui *duit.DUI, self *duit.Kid) (warp *image.Point) {
	p := ui.size.Div(2)
	return &p
}

func (ui *square) Focus(dui *duit.DUI, self *duit.Kid, o duit.UI) (warp *image.Point) {
	if ui != o {
		return nil
	}
	return ui.FirstFocus(dui, self)
}

func (ui *square) Mark(self *duit.Kid, o duit.UI, forLayout bool) (marked bool) {
	return self.Mark(o, forLayout)
}

func (ui *square) Print(self *duit.Kid, indent int) {
	duit.PrintUI(fmt.Sprintf("square dirty=%v", ui.dirty), self, indent)
}
