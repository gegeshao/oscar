package out

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mono83/oscar"
	"github.com/mono83/oscar/events"
	"io"
	"sync"
	"time"
)

// BuildDotRealTimePrinter returns events receiver, used to print test case flow
func BuildDotRealTimePrinter(stream io.Writer, enterAndLeave bool, total int) func(*events.Emitted) {
	cnt := 0
	max := 60
	m := sync.Mutex{}

	startedAt := time.Now()
	finished := 0

	print := func(s rune, c *color.Color) {
		str := string(s)
		if c != nil {
			str = c.Sprint(str)
		}

		m.Lock()
		fmt.Fprint(stream, str)
		cnt++
		if cnt == max {
			elapsed := time.Now().Sub(startedAt)
			if total < 1 || elapsed.Seconds() < 0 {
				fmt.Fprintf(stream, " Elapsed: %.1fs\n", elapsed.Seconds())
			} else {
				percent := 100. * float64(finished) / float64(total)
				if percent < 1 {
					percent = 1
				}
				eta := time.Duration(elapsed.Nanoseconds()*100/int64(percent) - elapsed.Nanoseconds())
				fmt.Fprintf(
					stream,
					" %.0f%% Elapsed: %.1fs ETA: ≈%.0fs\n",
					percent,
					elapsed.Seconds(),
					eta.Seconds(),
				)
			}

			cnt = 0
		}
		m.Unlock()
	}

	switcher := events.EventRouter{
		Assert: func(done events.AssertDone, _ *events.Emitted) {
			if done.Error == nil {
				print('.', colorDotOK)
			} else if oscar.IsSkip(done.Error) {
				print('s', colorDotSkip)
			} else {
				print('E', colorDotErr)
			}
		},
		Start: func(events.Start, *events.Emitted) {
			if enterAndLeave {
				print('<', colorDotSF)
			}
		},
		Finish: func(events.Finish, *events.Emitted) {
			if enterAndLeave {
				print('>', colorDotSF)
			}
			m.Lock()
			finished++
			m.Unlock()
		},
		Sleep: func(events.Sleep, *events.Emitted) {
			print('z', colorDotSleep)
		},
		Remote: func(events.RemoteRequest, *events.Emitted) {
			print('^', colorDotRemote)
		},
	}

	return switcher.OnEvent
}

var colorDotSF = color.New(color.FgBlack)
var colorDotOK = color.New(color.FgHiGreen)
var colorDotErr = color.New(color.FgRed)
var colorDotSleep = color.New(color.FgGreen)
var colorDotRemote = color.New(color.FgGreen)
var colorDotSkip = color.New(color.FgHiMagenta)
