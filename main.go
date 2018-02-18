// Acme & vi crossover editor, created with duit.
package main

import (
	"flag"
	"image"
	"log"
	"os"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

var (
	dui                                                   *duit.DUI
	topUI                                                 *mainUI
	squareDirtyColor, squareBorderColor, squareCleanColor *draw.Image
	tagColors, textColors                                 *duit.EditColors
)

const (
	control = 0x1f
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		log.Printf("usage: acvi [flags] [file ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()

	var err error
	dui, err = duit.NewDUI("acvi", &duit.DUIOpts{FontName: os.Getenv("font")})
	if err != nil {
		log.Fatalf("new dui: %s\n", err)
	}

	allocColor := func(v draw.Color) *draw.Image {
		c, err := dui.Display.AllocImage(image.Rect(0, 0, 1, 1), draw.ARGB32, true, v)
		if err != nil {
			log.Fatalf("alloccolor: %s\n", err)
		}
		return c
	}

	tagColors = &duit.EditColors{
		Fg:             allocColor(0x111111ff),
		Bg:             allocColor(0xeaffffff),
		SelFg:          dui.Regular.Normal.Text,
		SelBg:          allocColor(0x9eefeeff),
		ScrollVis:      allocColor(0xeaffffff),
		ScrollBg:       allocColor(0x4b9999ff), // same s,v offset as for textColor
		HoverScrollVis: allocColor(0xeaffffff),
		HoverScrollBg:  allocColor(0x3e8080ff), // -10 v
		CommandBorder:  dui.CommandMode,
		VisualBorder:   dui.VisualMode,
	}
	textColors = &duit.EditColors{
		Fg:             allocColor(0x111111ff),
		Bg:             allocColor(0xfffeeaff),
		SelFg:          dui.Regular.Normal.Text,
		SelBg:          allocColor(0xeeef9fff),
		ScrollVis:      allocColor(0xfffeeaff),
		ScrollBg:       allocColor(0x9a984bff),
		HoverScrollVis: allocColor(0xfffeeaff),
		HoverScrollBg:  allocColor(0x807e3eff), // -10 v
		CommandBorder:  dui.CommandMode,
		VisualBorder:   dui.VisualMode,
	}

	squareDirtyColor = allocColor(0x0e0098ff)
	squareBorderColor = allocColor(0x8888ccff)
	squareCleanColor = tagColors.Bg

	topUI = newMainUI(args)
	dui.Top.UI = topUI
	dui.Top.ID = "columns"

	for {
		select {
		case e := <-dui.Inputs:
			dui.Input(e)

		case err, ok := <-dui.Error:
			if !ok {
				return
			}
			log.Printf("duit: %s\n", err)
		}
	}
}
