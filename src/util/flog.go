package util

import (
	"time"
	"fmt"
	"io"
	"os"
)

type Area struct {
	Text string
	id uint64
}

type Flogger struct {
	File io.Writer
	// Area;[(AD|Area)...];EOA;Text;EOL
	Areas []Area
	AreaDivision string // AD
	EndOfArea string // EOA
	EndOfLine string // EOL
}

var StdoutFLog *Flogger
var Fl_ERROR uint64
func init() {
	_StdoutFLog := NewLogger(os.Stdout)
	StdoutFLog = &_StdoutFLog
	Fl_ERROR = StdoutFLog.NewArea("\x1b[0;31mError")
}

func FLog(areas uint64, format string, stuff... any) (n int, err error) {
	return StdoutFLog.Printf(areas, format, stuff...)
}

// WARNING: FlogSection (and FlogSection_m) don't implement io.Writer as it
// should Write(p []byte) (n int, err error) err is returned as it should, but
// it writes more than what is sent, including areas and time formatting. However it
// reports, falsely, it wrote exactly what the user sent
type FlogSection struct {
	SelectedAreas uint64
	Flog Flogger
}

type FlogSection_m struct {
	Prefix string
	Flog Flogger
}

func NewLogger(out io.Writer) Flogger {
	return Flogger{
		out, []Area{},
		"\x1b[0m ", // AD
		"\x1b[0m", // EOA
		"\x1b[0m", // EOL
	}
}

func (fl *Flogger) NewArea(text string) (id uint64) {
	id = 1<<uint64(len(fl.Areas))
	fl.Areas = append(fl.Areas, Area{text, id})
	return id
}

func (fl Flogger) Printf(areas uint64, format string, stuff ...any) (n int, err error) {
	var preamble string

	var i uint64
	for i=0;i<uint64(len(fl.Areas));i++ {
		if ((1<<i) & areas != 0) {
			area:=fl.Areas[i]
			if (preamble != "") {
				preamble+=fl.AreaDivision
			}
			preamble+=area.Text
		}
	}
	preamble+=fl.EndOfArea

	var text string
	if (format != "") {
		text = fmt.Sprintf(format, stuff...)
	}

	fmttime := time.Now().Format("02/01 15:04:05")
	return fmt.Fprintf(fl.File, "[%v] %s: %s%s", fmttime, preamble, text, fl.EndOfLine)
}

func (fls FlogSection) Write(p []byte) (n int, err error) {
	_, err = fls.Flog.Printf(fls.SelectedAreas, "%s", p)
	return len(p), err
}

func (fls_m FlogSection_m) Write(p []byte) (n int, err error) {
	fmttime := time.Now().Format("02/01 15:04:05")
	_, err = fmt.Fprintf(fls_m.Flog.File, "[%v] %s: %s%s", fmttime, fls_m.Prefix, p, fls_m.Flog.EndOfLine)
	return len(p), err
}

func (fl Flogger) Writer(areas uint64) io.Writer {
	return FlogSection{
		areas, fl,
	}
}

func (fl Flogger) Writer_m(prefix string) io.Writer {
	return FlogSection_m{
		prefix, fl,
	}
}

