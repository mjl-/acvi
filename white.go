package main

import (
	"image"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type white struct {
}

var _ duit.UI = &white{}

func (ui *white) Layout(dui *duit.DUI, self *duit.Kid, sizeAvail image.Point, force bool) {
	self.R = image.Rect(0, 0, sizeAvail.X, sizeAvail.Y)
}

func (ui *white) Draw(dui *duit.DUI, self *duit.Kid, img *draw.Image, orig image.Point, m draw.Mouse, force bool) {
	img.Draw(self.R.Add(orig), dui.Display.White, nil, image.ZP)
}

func (ui *white) Mouse(dui *duit.DUI, self *duit.Kid, m draw.Mouse, origM draw.Mouse, orig image.Point) (r duit.Result) {
	r.Hit = ui
	return
}

func (ui *white) Key(dui *duit.DUI, self *duit.Kid, k rune, m draw.Mouse, orig image.Point) (r duit.Result) {
	r.Hit = ui
	return
}

func (ui *white) FirstFocus(dui *duit.DUI, self *duit.Kid) (warp *image.Point) {
	p := self.R.Size().Div(2)
	return &p
}

func (ui *white) Focus(dui *duit.DUI, self *duit.Kid, o duit.UI) (warp *image.Point) {
	if ui != o {
		return nil
	}
	return ui.FirstFocus(dui, self)
}

func (ui *white) Mark(self *duit.Kid, o duit.UI, forLayout bool) (marked bool) {
	return self.Mark(o, forLayout)
}

func (ui *white) Print(self *duit.Kid, indent int) {
	duit.PrintUI("white", self, indent)
}
