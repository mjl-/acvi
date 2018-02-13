package main

import (
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type mainUI struct {
	columns []*columnUI
	duit.Split
}

func newMainUI(args []string) *mainUI {
	ui := &mainUI{}
	ui.Split.Gutter = 1
	ui.Split.Background = textColors.Fg
	if len(args) == 0 {
		dir, _ := os.Getwd()
		if dir != "" {
			args = []string{dir + "/"}
		}
		ui.columns = []*columnUI{
			newColumnUI(nil),
			newColumnUI(args),
		}
	} else {
		var args0, args1 []string
		if len(args) > 0 {
			args0 = args[:1]
			args1 = args[1:]
		}
		ui.columns = []*columnUI{
			newColumnUI(args0),
			newColumnUI(args1),
		}
	}
	uis := []duit.UI{
		ui.columns[0],
		ui.columns[1],
	}
	ui.Kids = duit.NewKids(uis...)
	return ui
}

func (ui *mainUI) error(filename string, err error, msg string) bool {
	if err == nil {
		return false
	}
	f := ui.ensureFile(errorDest(filename))
	f.append([]byte(fmt.Sprintf("%s: %s\n", msg, err)))
	return true
}

func (ui *mainUI) ensureFile(filename string) *fileUI {
	f := ui.findFile(filename)
	if f == nil {
		f = ui.columns[len(ui.columns)-1].addFile(filename)
	}
	return f
}

func (ui *mainUI) output(filename string, buf []byte) {
	ui.ensureFile(filename).append(buf)
}

func (ui *mainUI) findFile(filename string) *fileUI {
	for _, col := range ui.columns {
		for _, file := range col.files.files {
			if file.path() == filename {
				return file
			}
		}
	}
	return nil
}

var addressRegexp *regexp.Regexp

func init() {
	addressRegexp = regexp.MustCompile(`^([0-9]+)(:([0-9]+))?:?$`)
}

func (ui *mainUI) look(filename, s string, focus bool) (consumed bool) {
	// todo: try plumber

	log.Printf("look %q %q\n", filename, s)

	p := s
	if filename != "" && !path.IsAbs(p) {
		p = path.Join(path.Dir(filename), s)
	}

	t := strings.SplitN(p, ":", 2)
	p = t[0]

	selectAddress := func(f *fileUI) (match bool) {
		if len(t) != 2 {
			return
		}
		addr := t[1]
		if strings.HasPrefix(addr, "/") {
			f.body.LastSearch = addr
			match = f.body.Search(dui, false)
			return
		}

		l := addressRegexp.FindStringSubmatch(addr)
		if l == nil {
			topUI.error(filename, fmt.Errorf("bad address"), "parsing address")
			return
		}
		lines, err := strconv.ParseInt(l[1], 10, 64)
		if topUI.error(filename, err, "parsing linenumber in address") {
			return
		}
		fr := f.body.EditReader(0)
		for ; lines > 1; lines-- {
			fr.Line(true)
		}
		var c duit.Cursor
		if l[3] != "" {
			lineoffset, err := strconv.ParseInt(l[3], 10, 64)
			if topUI.error(filename, err, "parsing line offset in address") {
				return
			}
			for ; lineoffset > 1; lineoffset-- {
				fr.TryGet()
			}
			c = duit.Cursor{Cur: fr.Offset(), Start: fr.Offset()}
		} else {
			c.Start = fr.Offset()
			fr.Line(true)
			c.Cur = fr.Offset()
		}
		f.body.SetCursor(c)
		f.body.ScrollCursor(dui)
		match = true
		return
	}

	f := ui.findFile(p)
	if f != nil {
		selectAddress(f)
		if focus {
			dui.Focus(f)
		}
		return true
	}

	for _, col := range ui.columns {
		for _, file := range col.files.files {
			if file.path() == p {
				selectAddress(file)
				if focus {
					dui.Focus(file)
				}
				return true
			}
		}
	}

	info, err := os.Stat(p)
	if err == nil {
		// todo: find spot with biggest area, split it, preferrably in another column
		if info.IsDir() && !strings.HasSuffix(p, "/") {
			p += "/"
		}
		file := ui.columns[0].addFile(p)
		selectAddress(file)
		return true
	}

	return false
}

func (ui *mainUI) execute(filename, cmd string, edit *duit.Edit) {
	switch cmd {
	case "Newcol":
		col := newColumnUI(nil)
		ui.columns = append(ui.columns, col)
		ui.Kids = append(ui.Kids, &duit.Kid{UI: col})
		dui.MarkLayout(ui)
	case "Exit":
		log.Printf("exit\n")
		dui.Close()
		os.Exit(0)
	case "Open":
		if edit == nil {
			topUI.error(filename, fmt.Errorf("needs selection"), "open files")
			return
		}
		sel, err := edit.Selection()
		if ui.error(filename, err, "selection") {
			return
		}
		l := strings.Split(string(sel), "\n")
		for _, name := range l {
			ui.look(filename, name, false)
		}
	default:
		if filename == "" {
			filename, _ = os.Getwd()
		}
		dest := errorDest(filename)
		cmd = strings.TrimSpace(cmd)
		if strings.HasPrefix(cmd, "|") {
			what := strings.Split(cmd[1:], " ")[0]
			if edit == nil {
				topUI.error(filename, fmt.Errorf("only works on files"), "execute filter")
				return
			}
			cc := edit.Cursor()
			sel, err := edit.Selection()
			if ui.error(filename, err, "selection") {
				return
			}
			go func() {
				c := exec.Command("sh", "-c", cmd[1:])
				stdin, err := c.StdinPipe()
				var buf []byte
				if err == nil {
					go func() {
						_, err := stdin.Write([]byte(sel))
						err2 := stdin.Close()
						if err == nil {
							err = err2
						}
						if err != nil {
							dui.Call <- func() {
								topUI.error(filename, err, "write to stdin of "+what)
							}
						}
					}()
					buf, err = c.Output()
				}
				dui.Call <- func() {
					if topUI.error(filename, err, what) {
						return
					}
					edit.Replace(cc, buf)
					dui.MarkDraw(edit)
				}
			}()
			return
		}
		go func() {
			what := strings.Split(cmd, " ")[0]
			c := exec.Command("sh", "-c", cmd)
			p, err := c.StdoutPipe()
			if err != nil {
				dui.Call <- func() {
					ui.error(filename, err, what+": stdoutpipe")
				}
				return
			}

			c.Stderr = c.Stdout
			err = c.Start()
			if err != nil {
				p.Close()
				dui.Call <- func() {
					ui.error(filename, err, what+": start")
				}
				return
			}

			done := make(chan struct{}, 1)
			for {
				buf := make([]byte, 8*1024)
				n, err := p.Read(buf)
				if n > 0 {
					dui.Call <- func() {
						topUI.output(dest, buf[:n])
						done <- struct{}{}
					}
				}
				if err == io.EOF {
					break
				}
				if ui.error(filename, err, what+": read") {
					break
				}
				<-done
			}
			err = c.Wait()
			if err != nil {
				dui.Call <- func() {
					ui.error(filename, err, what)
				}
			}
		}()
	}
}

func (ui *mainUI) removeColumn(col *columnUI) {
	if len(ui.columns) == 1 {
		return
	}
	for i := range ui.columns {
		if ui.columns[i] == col {
			copy(ui.columns[i:], ui.columns[i+1:])
			copy(ui.Kids[i:], ui.Kids[i+1:])
			ui.columns = ui.columns[:len(ui.columns)-1]
			ui.Kids = ui.Kids[:len(ui.Kids)-1]
			return
		}
	}
}

func (ui *mainUI) complete(filename string, edit *duit.Edit) {
	tt, err := edit.Selection()
	if ui.error(filename, err, "selection") {
		return
	}
	_, c := edit.Cursor().Ordered()
	t := string(tt)
	if t == "" {
		r := edit.ReverseEditReader(c)
		for {
			c, eof := r.Peek()
			if eof || c <= ',' {
				break
			}
			r.Get()
			t = string(c) + t
		}
	}
	if !strings.HasPrefix(t, "/") {
		t = path.Dir(filename) + "/" + t
	}
	// log.Printf("full path %q, dir %q, base %q\n", t, path.Dir(t), path.Base(t))
	files, err := ioutil.ReadDir(path.Dir(t))
	if topUI.error(filename, err, "readdir") {
		return
	}
	name := path.Base(t)
	var matches []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), name) {
			matches = append(matches, f.Name())
		}
	}
	if len(matches) == 0 {
		topUI.error(filename, fmt.Errorf("no matches"), fmt.Sprintf("completing %q", t))
		return
	}
	if len(matches) > 1 {
		m0 := matches[0]
		prefixLen := len(m0)
		for i := 1; i < len(matches); i++ {
			m := matches[i]
			mlen := len(m)
			for j := 0; j < prefixLen && j < mlen; j++ {
				if m0[j] != m[j] {
					prefixLen = j
					break
				}
			}
		}
		if prefixLen <= len(name) {
			s := "\n" + strings.Join(matches, "\n")
			topUI.error(filename, fmt.Errorf("multiple matches:%s", s), fmt.Sprintf("completing %q", t))
			return
		}
		matches = []string{m0[:prefixLen]}
	}
	add := matches[0][len(name):]
	edit.Replace(duit.Cursor{Cur: c, Start: c}, []byte(add))
	c += int64(len(add))
	edit.SetCursor(duit.Cursor{Cur: c, Start: c})
	dui.MarkDraw(edit)
}

func (ui *mainUI) mouseColumn(m draw.Mouse) int {
	for i, col := range ui.Kids {
		if m.X >= col.R.Min.X && m.X < col.R.Max.X {
			return i
		}
	}
	return -1
}

func (ui *mainUI) columnIndex(col *columnUI) int {
	for i, c := range ui.columns {
		if c == col {
			return i
		}
	}
	return -1
}

func (ui *mainUI) grow(col *columnUI) {
	i := ui.columnIndex(col)
	if i >= 0 {
		ui.growColumnIndex(i)
	}
}

func (ui *mainUI) growAvailable(col *columnUI) {
	i := ui.columnIndex(col)
	if i < 0 {
		return
	}
	dims := ui.Split.Dimensions(dui, nil)
	total := 0
	for _, d := range dims {
		total += d
	}
	scrollbarSize := dui.Scale(duit.ScrollbarSize)
	d := total - (len(dims)-1)*scrollbarSize
	for j := range dims {
		if j == i {
			dims[j] = d
		} else {
			dims[j] = scrollbarSize
		}
	}
	setMinDims(dims, scrollbarSize)
	ui.Dimensions(dui, dims)
	dui.MarkLayout(ui)
}

func (ui *mainUI) growFull(col *columnUI) {
	i := ui.columnIndex(col)
	if i < 0 {
		return
	}
	dims := ui.Split.Dimensions(dui, nil)
	total := 0
	for _, d := range dims {
		total += d
	}
	for j := range dims {
		if j == i {
			dims[j] = total
		} else {
			dims[j] = 0
		}
	}
	ui.Dimensions(dui, dims)
	dui.MarkLayout(ui)
}

func (ui *mainUI) growColumnIndex(i int) {
	dims := ui.Split.Dimensions(dui, nil)
	scrollbarSize := dui.Scale(duit.ScrollbarSize)
	grow(dui.Scale(100), i, dims, scrollbarSize)
	setMinDims(dims, scrollbarSize)
	ui.Dimensions(dui, dims)
	dui.MarkLayout(ui)
}

func (ui *mainUI) Key(dui *duit.DUI, self *duit.Kid, k rune, m draw.Mouse, orig image.Point) (r duit.Result) {
	switch k {
	case draw.KeyCmd + 'H':
		i := (ui.mouseColumn(m) - 1 + len(ui.Kids)) % len(ui.Kids)
		p := image.Pt(ui.Kids[i].R.Min.X+ui.Kids[i].R.Dx()/2, m.Y).Add(orig)
		r.Warp = &p
	case draw.KeyCmd + 'L':
		i := (ui.mouseColumn(m) + 1) % len(ui.Kids)
		p := image.Pt(ui.Kids[i].R.Min.X+ui.Kids[i].R.Dx()/2, m.Y).Add(orig)
		r.Warp = &p
	case draw.KeyCmd + 'i':
		i := ui.mouseColumn(m)
		if i < 0 {
			return
		}
		ui.growColumnIndex(i)
	case draw.KeyCmd + '1', draw.KeyCmd + '2', draw.KeyCmd + '3':
		m.Buttons = 1 << uint(k-(draw.KeyCmd+'1'))
		r0 := ui.Mouse(dui, self, m, m, orig)
		m.Buttons = 0
		r = ui.Mouse(dui, self, m, m, orig)
		if r.Warp == nil && r0.Warp != nil {
			r.Warp = r0.Warp
		}
		r.Consumed = r.Consumed || r0.Consumed
	default:
		return ui.Split.Key(dui, self, k, m, orig)
	}
	r.Consumed = true
	return
}

func (ui *mainUI) Print(self *duit.Kid, indent int) {
	duit.PrintUI("mainUI", self, indent)
	duit.KidsPrint(ui.Kids, indent+1)
}
