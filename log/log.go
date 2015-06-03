package log

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	level int
)

func init() {
	lvl := os.Getenv("LOGLEVEL")
	if lvl != "" {
		level = code(lvl)
	} else {
		level = code("info")
	}
}

func code(lvl string) int {
	switch strings.ToLower(lvl) {
	case "all":
		return 0
	case "trace":
		return 10
	case "debug":
		return 20
	case "error", "err":
		return 30
	case "fatal", "crit", "critical":
		return 40
	case "info":
		return 50
	case "warn", "warning":
		return 60
	case "off":
		return 100
	default:
		return 50
	}
}

func log(lvl, msg string) {
	c := code(lvl)
	if level <= c {
		ts := time.Now().Format(time.RFC3339)
		if c < 50 {
			fmt.Fprintf(os.Stderr, "%s rdpg-agent %s %s\n", ts, lvl, msg)
		} else {
			// TODO: Standard output logs via channel.
			fmt.Fprintf(os.Stdout, "%s rdpg-agent %s %s\n", ts, lvl, msg)
		}
	}
}

func Trace(msg string) {
	log("trace", msg)
}

func Debug(msg string) {
	log("debug", msg)
}

func Error(msg string) {
	log("error", msg)
}

func Fatal(msg string) {
	log("fatal", msg)
}

func Info(msg string) {
	log("info", msg)
}

func Warn(msg string) {
	log("warn", msg)
}
