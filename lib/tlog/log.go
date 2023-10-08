package tlog

import (
	"encoding/json"
	"fmt"
	"lib/terrors"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level string

const (
	LevelDebug   Level = "DEBUG"
	LevelInfo    Level = "INFO"
	LevelWarning Level = "WARN"
	LevelError   Level = "ERROR"
	LevelFatal   Level = "FATAL"
)

type TraceS struct {
	FilePath     string
	LineNr       int
	FunctionName string
}

type RecordS struct {
	// ContextID string        `yaml:"context-id"`
	Level     Level         `yaml:"level"`
	Message   string        `yaml:"message"`
	Time      string        `yaml:"time"`
	Vars      []interface{} `yaml:"vars"`
	Traceback []TraceS      `yaml:"traceback"`
}

type entryDebugS struct {
	Log       map[string]string
	Traceback []string
}

type Vars map[string]interface{}

func Debug(in ...interface{}) *RecordS { return entryCreate(LevelDebug, in) }

func Info(in ...interface{}) *RecordS { return entryCreate(LevelInfo, in) }

func Warning(in ...interface{}) *RecordS { return entryCreate(LevelWarning, in) }

func Error(in ...interface{}) *RecordS { return entryCreate(LevelError, in) }

// Fatal ...
func Fatal(in ...interface{}) *RecordS {
	if entryCreate(LevelFatal, in) != nil {
		os.Exit(1)
	}
	return nil
}

var framesToIgnore = []string{
	"runtime.goexit",
	"runtime.main",
	"net/http.(*conn).serve",
	"net/http.(*ServeMux).ServeHTTP",
	"net/http.HandlerFunc.ServeHTTP",
	"net/http.serverHandler.ServeHTTP",
}

func entryCreate(level Level, in []interface{}) *RecordS {

	if level == LevelDebug && os.Getenv("LOG_SHOW_DEBUG") == "" {
		return nil
	}

	inLen := len(in)
	if inLen == 0 {
		return nil
	}
	if in[0] == nil || in[0] == "" {
		return nil
	}

	switch in[0].(type) {
	case *RecordS:
		return in[0].(*RecordS)

	case terrors.Error:
		in[0] = fmt.Sprintf("$c:%d", in[0])
	}

	// ----------------------------------------------------
	// creating new log entry

	// goID := goroutineID()
	entry := &RecordS{
		// ID:        goID + "-" + common.RandString(8) + "-" + string(level[0]),
		// ContextID: goID,
		Level:     level,
		Message:   fmt.Sprint(in[0]),
		Time:      time.Now().Format("Jan 2 15:04:05.000000"),
		Vars:      in[1:],
		Traceback: Trace(),
	}

	// ----------------------------------------------------
	// print JSON to stdout

	out := map[string]string{
		// "context-id": entry.ContextID,
		"time":    entry.Time,
		"level":   string(entry.Level),
		"message": entry.Message,
	}

	traceback := []string{}
	for _, v := range entry.Traceback {
		traceback = append(traceback, fmt.Sprintf("%s:%d", v.FilePath, v.LineNr))
	}

	for i, v := range entry.Vars {
		switch vt := v.(type) {
		case string:
			out[fmt.Sprintf("var/%d", i)] = vt

		case Vars:
			for k2, v2 := range vt {
				if k2 == "context-id" {
					k2 = "var/context-id"
				}
				if k2 == "level" {
					k2 = "var/level"
				}
				if k2 == "message" {
					k2 = "var/message"
				}
				if k2 == "traceback" {
					k2 = "var/traceback"
				}
				out[k2] = fmt.Sprint(v2)
			}

		default:
			out[fmt.Sprintf("var/%d", i)] = fmt.Sprint(v)
		}
	}

	// ----------------------------------------------------
	// message template

	if strings.Contains(entry.Message, "{{") {
		for k, v := range out {
			entry.Message = strings.ReplaceAll(entry.Message, "{{"+k+"}}", v)
		}
		out["message"] = entry.Message
	}

	// ----------------------------------------------------

	switch strings.TrimSpace(strings.ToLower(os.Getenv("LOG_MODE"))) {
	case "simple", "oneline":

		fmt.Println(time.Now().String()[:23], entry.Level[:1], " ", entry.Message)

	case "multiline-json", "mjson":
		// print multi-line JSON
		buf, _ := json.MarshalIndent(entryDebugS{
			Log:       out,
			Traceback: traceback,
		}, "", "  ")
		fmt.Println(string(buf))

	default:

		out["traceback"] = strings.Join(traceback, "\n")
		buf, _ := json.Marshal(out)
		fmt.Println(string(buf))

	}

	// ----------------------------------------------------

	return entry
}

// --------------------------------------------------

func Trace() []TraceS {
	trace := []TraceS{}
	pc := make([]uintptr, 40)
	n := runtime.Callers(4, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, isMore := frames.Next()
		if !isStringInSlice(frame.Function, framesToIgnore) {
			trace = append(trace, TraceS{
				FilePath:     frame.File,
				LineNr:       frame.Line,
				FunctionName: frame.Function,
			})
		}

		if !isMore {
			break
		}
	}
	return trace
}

func Trace3() []TraceS {
	trace := []TraceS{}
	pc := make([]uintptr, 40)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, isMore := frames.Next()
		if !isStringInSlice(frame.Function, framesToIgnore) {
			trace = append(trace, TraceS{
				FilePath:     frame.File,
				LineNr:       frame.Line,
				FunctionName: frame.Function,
			})
		}

		if !isMore {
			break
		}
	}
	return trace
}

func isStringInSlice(s string, slice []string) bool {
	for _, ss := range slice {
		if ss == s {
			return true
		}
	}
	return false
}

// --------------------------------------------------

func PrintJSON(in interface{}) error {

	if os.Getenv("LOG_SHOW_DEBUG") != "" {
		fmt.Println(Trace3()[0])
	}

	// ---

	buf, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(buf))
	return nil
}
