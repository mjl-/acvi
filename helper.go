package main

import (
	"io"
	"log"
	"path"
	"strings"

	"github.com/mjl-/duit"
)

func unscale(v int) int {
	return v * 1000 / dui.Scale(1000)
}

func buttonText(edit *duit.Edit) string {
	t, err := edit.Selection()
	topUI.error("", err, "button selection")
	if len(t) == 0 {
		t, err = edit.ExpandedText()
		topUI.error("", err, "button expanded text")
	}
	return string(t)
}

func readAtFull(src io.ReaderAt, buf []byte, offset int64) (read int, err error) {
	want := len(buf)
	for want > 0 {
		var n int
		n, err = src.ReadAt(buf, offset)
		if n > 0 {
			read += n
			offset += int64(n)
			want -= n
			buf = buf[n:]
		}
		if err == io.EOF {
			return
		}
	}
	return
}

func expandText(edit *duit.Edit, offset int64) string {
	c := edit.Cursor()
	c0, c1 := c.Ordered()
	if c0 != c1 && offset >= c0 && offset <= c1 {
		buf, err := edit.Selection()
		topUI.error("", err, "expanded text")
		return string(buf)
	}
	br := edit.ReverseEditReader(offset)
	fr := edit.EditReader(offset)
	br.Nonwhitespace()
	fr.Nonwhitespace()
	buf := make([]byte, int(fr.Offset()-br.Offset()))
	n, err := readAtFull(edit.Reader(), buf, br.Offset())
	if err != nil {
		log.Printf("read: %s\n", err)
		return ""
	}
	return string(buf[:n])
}

func errorDest(s string) string {
	if !strings.HasSuffix(s, "/+Errors") {
		if !strings.HasSuffix(s, "/") {
			s = path.Dir(s) + "/"
		}
		s += "+Errors"
	}
	return s
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maximum(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maximum64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func grow(ogrow int, index int, dims []int, minDim int) (taken int) {
	take := func(want int, dims []int, i, delta int) int {
		avail := 0
		for j := i + delta; j >= 0 && j < len(dims); j += delta {
			avail += maximum(0, dims[j]-minDim)
		}
		if want >= avail {
			for j := i + delta; j >= 0 && j < len(dims); j += delta {
				dims[j] = minDim
			}
			want -= avail
		} else if avail > 0 {
			stillNeed := want
			for j := i + delta; j >= 0 && j < len(dims); j += delta {
				if j == 0 || j == len(dims)-1 {
					dims[j] = maximum(minDim, dims[j]-stillNeed)
				} else {
					d := dims[j] - (maximum(0, dims[j]-minDim))*want/avail
					stillNeed -= dims[j] - d
					dims[j] = d
				}
			}
			want = 0
		}
		return want
	}
	grow := take(ogrow, dims, index, 1)
	grow = take(grow, dims, index, -1)
	taken = ogrow - grow
	dims[index] += taken
	return taken
}

func tagHeight() int {
	return 1 + unscale(dui.Display.DefaultFont.Height)
}

func setMinDims(dim []int, min int) {
	need := 0
	found := 0
	for _, d := range dim {
		if d > min {
			found += d - min
		} else {
			need += min - d
		}
	}
	found = minimum(found, need)
	for i, d := range dim {
		if d > min && need > 0 {
			take := minimum(need, d-min)
			need -= take
			dim[i] -= take
		} else if d < min && found > 0 {
			take := minimum(found, min-d)
			dim[i] += take
			found -= take
		}
	}
}
