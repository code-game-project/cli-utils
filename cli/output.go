package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Color string

const (
	Reset       Color = "\x1b[0m"
	Black       Color = "\x1b[30m"
	Red         Color = "\x1b[31m"
	Green       Color = "\x1b[32m"
	Yellow      Color = "\x1b[33m"
	Blue        Color = "\x1b[34m"
	Magenta     Color = "\x1b[35m"
	Cyan        Color = "\x1b[36m"
	White       Color = "\x1b[37m"
	BlackBold   Color = "\x1b[1;30m"
	RedBold     Color = "\x1b[1;31m"
	GreenBold   Color = "\x1b[1;32m"
	YellowBold  Color = "\x1b[1;33m"
	BlueBold    Color = "\x1b[1;34m"
	MagentaBold Color = "\x1b[1;35m"
	CyanBold    Color = "\x1b[1;36m"
	WhiteBold   Color = "\x1b[1;37m"
	WhiteDim    Color = "\x1b[2;37m"
)

var (
	out                = colorable.NewColorableStdout()
	progressStart      time.Time
	progressMsg        string
	loadingTicker      *time.Ticker
	progressBarRunning bool
)

type Unit int

const (
	UnitNone     Unit = 0
	UnitFileSize Unit = decor.UnitKiB
)

type progressBar struct {
	bar    *mpb.Bar
	update chan any
}

var progressBars = make(map[string]progressBar)

func Progress(key, message string, current, total int64, unit Unit) {
	if b, ok := progressBars[key]; ok {
		b.bar.SetCurrent(current)
		b.update <- struct{}{}
		if current == total {
			b.update <- struct{}{}
			b.bar.Wait()
			b.update <- struct{}{}
			delete(progressBars, key)
		}
	} else {
		b := progressBar{
			update: make(chan any),
		}
		p := mpb.New(mpb.WithWidth(64), mpb.WithManualRefresh(b.update))

		decorators := make([]decor.Decorator, 0, 1)
		if unit == UnitFileSize {
			decorators = append(decorators, decor.Counters(int(unit), "%.1f / %.1f"))
		}
		decorators = append(decorators, decor.NewPercentage("  %.2f"))

		b.bar = p.New(int64(total),
			mpb.BarStyle(),
			mpb.PrependDecorators(decor.Name(message)),
			mpb.AppendDecorators(decorators...),
		)
		progressBars[key] = b
	}
}

func CancelProgressBars() {
	for _, b := range progressBars {
		b.bar.Abort(false)
	}
	progressBars = make(map[string]progressBar)
}

func Print(format string, a ...any) {
	fmt.Fprintf(out, "%s\n", fmt.Sprintf(format, a...))
}

func PrintColor(color Color, format string, a ...any) {
	fmt.Fprintf(out, "%s%s%s\n", color, fmt.Sprintf(format, a...), Reset)
}

func Success(format string, a ...any) {
	PrintColor(Green, format, a...)
}

func Warn(format string, a ...any) {
	Print(string(Yellow)+"WARNING: "+string(Reset)+format, a...)
}

func Error(format string, a ...any) {
	Print(string(RedBold)+"ERROR: "+string(Reset)+format, a...)
}

func Clear() {
	fmt.Fprintf(out, "\033[H\033[2J")
}

func SetColor(color Color) {
	fmt.Fprint(out, color)
}

func ResetColor() {
	fmt.Fprint(out, Reset)
}

func Output() io.Writer {
	return out
}
