package glog

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type SlogReplacerOpts struct {
	AddSource bool
	AddTime   TimeFormat
}

type TimeFormat uint8

const (
	TimeFormatNone     = 0
	TimeFormatDuration = 1
	TimeFormatRFC3339  = 2
)

func SlogReplacer(opts SlogReplacerOpts) func(groups []string, a slog.Attr) slog.Attr {
	initialTime := time.Now()
	return func(groups []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			if opts.AddTime == TimeFormatNone {
				break
			}
			var orig time.Time
			switch a.Value.Kind() {
			case slog.KindTime:
				orig = a.Value.Time()
			case slog.KindString:
				t, err := time.Parse(time.RFC3339, a.Value.String())
				if err == nil {
					orig = t
					break
				}
				t, err = time.Parse(time.RFC3339Nano, a.Value.String())
				if err == nil {
					orig = t
					break
				}
			case slog.KindUint64:
				orig = time.Unix(int64(a.Value.Uint64()), 0)
			case slog.KindInt64:
				orig = time.Unix(a.Value.Int64(), 0)
				// case slog.KindFloat64:
				// 	integer := a.Value.Float64()
			}
			switch opts.AddTime {
			case TimeFormatDuration:
				dur := orig.Sub(initialTime)
				a.Value = slog.DurationValue(dur.Round(time.Microsecond))
			case TimeFormatRFC3339:
				a.Value = slog.TimeValue(orig)
			}

		case slog.SourceKey:
			if !opts.AddSource {
				break
			}
			src, ok := a.Value.Any().(*slog.Source)
			if !ok {
				break
			}
			cwd, err := os.Getwd()
			if err != nil {
				break
			}
			gopath := os.Getenv("GOPATH")
			fmt.Println("src", src)
			for _, prefix := range []string{gopath, cwd} {
				if smaller, ok := strings.CutPrefix(src.File, prefix); ok {
					src.File = smaller
				}
			}
		default:
		}
		return a
	}
}

type PflagLeveler struct{ Flag *string }

var _ slog.Leveler = PflagLeveler{}

func (l PflagLeveler) Level() slog.Level {
	if l.Flag == nil {
		return slog.LevelInfo
	}
	s := *l.Flag
	return ParseLevel(s)
}
