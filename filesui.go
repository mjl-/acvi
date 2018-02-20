package main

import (
	"image"

	"github.com/mjl-/duit"
)

type filesUI struct {
	column  *columnUI
	files   []*fileUI
	heights []int
	duit.Box
}

func newFilesUI(column *columnUI, files []*fileUI) *filesUI {
	ui := &filesUI{
		column: column,
		files:  files,
	}
	uis := make([]duit.UI, len(files))
	for i, f := range files {
		uis[i] = f
	}

	ui.Box.Margin = image.Pt(0, 2) // todo: fix the off-by one, or rather, don't use box to layout, doesn't make enough sense
	ui.Box.Background = textColors.Fg
	ui.Box.Width = -1
	if len(uis) == 0 {
		ui.Box.Kids = duit.NewKids(&white{})
	} else {
		ui.Box.Kids = duit.NewKids(uis...)
	}
	return ui
}

func (ui *filesUI) marginY() int {
	return dui.Scale(ui.Box.Margin.Y)
}

// total heights excluding padding
func (ui *filesUI) height() (total int) {
	for _, h := range ui.heights {
		total += h
	}
	return
}

func (ui *filesUI) add(file *fileUI) {
	if len(ui.files) == 0 {
		ui.files = append(ui.files, file)
		ui.Box.Kids = duit.NewKids(file) // cannot append, kids currently has a "white"
		ui.heights = append(ui.heights, ui.height())
	} else {
		var lgi, lgh int
		for i, h := range ui.heights {
			if i == 0 || h >= lgh {
				lgi, lgh = i, h
			}
		}
		// nopes, not pretty
		ui.files = append(append(append([]*fileUI{}, ui.files[:lgi+1]...), file), ui.files[lgi+1:]...)
		ui.Kids = append(append(append([]*duit.Kid{}, ui.Kids[:lgi+1]...), &duit.Kid{UI: file}), ui.Kids[lgi+1:]...)
		ui.heights = append(append(append([]int{}, ui.heights[:lgi]...), ui.heights[lgi]/2-ui.marginY(), ui.heights[lgi]-ui.heights[lgi]/2), ui.heights[lgi+1:]...)
	}
	dui.MarkLayout(nil)
	dui.Focus(file)
}

func (ui *filesUI) remove(file *fileUI) {
	for i := range ui.files {
		if ui.files[i] != file {
			continue
		}
		h := ui.heights[i]
		copy(ui.files[i:], ui.files[i+1:])
		copy(ui.Kids[i:], ui.Kids[i+1:])
		copy(ui.heights[i:], ui.heights[i+1:])
		ui.files = ui.files[:len(ui.files)-1]
		ui.Kids = ui.Kids[:len(ui.Kids)-1]
		ui.heights = ui.heights[:len(ui.heights)-1]
		if i < len(ui.heights) {
			ui.heights[i] += h + ui.marginY()
		} else if i-1 >= 0 && i-1 < len(ui.heights) {
			ui.heights[i-1] += h + ui.marginY()
		}
		if len(ui.files) == 0 {
			ui.Box.Kids = duit.NewKids(&white{})
		}
		dui.MarkLayout(nil)
		dui.Focus(file)
		return
	}
	panic("no file removed?")
}

func (ui *filesUI) Layout(dui *duit.DUI, self *duit.Kid, sizeAvail image.Point, force bool) {
	if duit.KidsLayout(dui, self, ui.Kids, force) {
		return
	}

	if len(ui.Kids) == 0 {
		return
	}

	if len(ui.heights) != len(ui.Kids) {
		// initial assignment
		ui.heights = make([]int, len(ui.Kids))
		single := (sizeAvail.Y - ui.marginY()*(len(ui.Kids)-1)) / len(ui.Kids)
		marginY := ui.marginY()
		for i := range ui.Kids {
			if i == len(ui.Kids)-1 {
				ui.heights[i] = sizeAvail.Y - i*(single+marginY)
			} else {
				ui.heights[i] = single
			}
		}
	} else {
		total := ui.height()
		if total != sizeAvail.Y {
			if total == 0 {
				if len(ui.heights) != 1 {
					panic("total height 0 with multiple files?")
				}
				ui.heights[0] = sizeAvail.Y
			} else {
				// probably new layout height, assign anew in same ratio
				left := sizeAvail.Y - ui.marginY()*(len(ui.Kids)-1)
				for i, h := range ui.heights {
					if i == len(ui.heights)-1 {
						ui.heights[i] = left
					} else {
						ui.heights[i] = h * sizeAvail.Y / total
					}
					left -= ui.heights[i]
				}
				if left != 0 {
					panic("left != 0")
				}
			}
		}
	}

	y := 0
	marginY := ui.marginY()
	for i, k := range ui.Kids {
		k.UI.Layout(dui, k, image.Pt(sizeAvail.X, ui.heights[i]), true)
		k.R = k.R.Add(image.Pt(0, y))
		y += ui.heights[i] + marginY
	}
	y--
	self.R = image.Rect(0, 0, sizeAvail.X, y)
}

func (ui *filesUI) Print(self *duit.Kid, indent int) {
	duit.PrintUI("filesUI", self, indent)
	duit.KidsPrint(ui.Kids, indent+1)
}
