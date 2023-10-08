package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"journal-proxy/global"
	"journal-proxy/parser/custom"
	"lib/utils/conv"
	"lib/utils/maps"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

var (
	replacer = strings.NewReplacer("\n", "", "\t", "", "  ", " ")

	parsers maps.Map[string, *custom.Parser]
)

type LevelT string

const (
	LevelDebug   LevelT = "DEBUG"
	LevelInfo    LevelT = "INFO"
	LevelWarning LevelT = "WARN"
	LevelError   LevelT = "ERROR"
	LevelFatal   LevelT = "FATAL"
)

func levelFromString(level string) LevelT {
	if level == "" {
		return LevelInfo
	}
	switch /* lowercase */ level[0] | ' ' {
	case 'd', 't':
		return LevelDebug
	case 'w':
		return LevelWarning
	case 'e':
		return LevelError
	case 'f':
		return LevelFatal
	default:
		return LevelInfo
	}
}

func Parse(message global.Message) []*global.Entry {
	if len(message.Data) == 0 {
		return nil
	}

	entries := []*global.Entry{}

	var customParser *custom.Parser
	if message.Parser != "" {
		parser, exists := parsers.GetFull(message.Parser)
		if !exists {
			parser = custom.NewParser(strings.NewReader(message.Parser))
			parsers.Set(message.Parser, parser)
		}
		customParser = parser
	}

	for _, lg := range splitLogs(message.Data) {
		if lg == "" {
			continue
		}

		entry := message.ToEntry()
		entry.Message = lg

		data := []byte(lg)
		switch {
		case json.Valid(data):
			unpackJSON(data, entry, "" /* ,entry.Element */, true)
			if entry.TagsString["event"] == "true" {
				entry.Event = true
			}

		case customParser != nil:
			// if not json, try custom parser
			customParser.Parse(entry)
		}

		fixLevel(entry)
		entry.EnvID = conv.String(entry.EnvID)
		entry.Time = uint64(time.Now().UnixNano())
		entries = append(entries, entry)
	}
	return entries
}

func fixLevel(entry *global.Entry) {
	entry.Level = string(levelFromString(entry.Level))

	for k, v := range entry.TagsNumber {
		if strings.HasSuffix(k, "status") {
			entry.Level = string(httpLevel(int(v)))
			return
		}
	}
}

func httpLevel(statusCode int) LevelT {
	switch {
	case statusCode >= 100 && statusCode < 200:
		// Informational
		return LevelInfo
	case statusCode >= 200 && statusCode < 300:
		// Success
		return LevelInfo
	case statusCode >= 300 && statusCode < 400:
		// Redirection
		return LevelInfo
	case statusCode >= 400 && statusCode < 500:
		// Client error
		return LevelWarning
	case statusCode >= 500 && statusCode < 600:
		// Server error
		return LevelError
	default:
		return LevelInfo
	}
}

func cleanLog(logs []string, data ...string) []string {
	for _, lg := range data {
		logs = append(logs, strings.TrimSpace(replacer.Replace(lg)))
	}
	return logs
}

func splitLogs(data []byte) (logs []string) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanRunes)

	idx := 0
	lastLog := 0
	openBrackets := 0

	for scanner.Scan() {
		idx++
		switch scanner.Text() {
		case "{":
			openBrackets++
			continue
		case "}":
			openBrackets--
		case "\n":

		default:
			continue
		}

		if openBrackets == 0 {
			logs = cleanLog(logs, string(data[lastLog:idx]))
			lastLog = idx
		}
	}

	// append last logs
	if lastLog < len(data) {
		logs = cleanLog(logs, strings.Split(string(data[lastLog:]), "\n")...)
	}

	// scanner error
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return logs
}

func unpackJSON(data []byte, entry *global.Entry, prefix string, topLevel bool) {
	jsonparser.ObjectEach(data, func(keyRaw, valueRaw []byte, valueType jsonparser.ValueType, _ int) error {
		keyStr := string(keyRaw)
		valueStr := string(valueRaw)

		val := struct {
			string
			float64
		}{}

		prefixStr := prefix + "/" + keyStr
		switch valueType {
		case jsonparser.Array:
			jsonparser.ArrayEach(valueRaw, func(data []byte, valueType jsonparser.ValueType, _ int, err error) {
				if err != nil {
					return
				}
				switch valueType {
				case jsonparser.Object:
					unpackJSON(data, entry, prefixStr, false)
				default:
					val.string = string(valueRaw)
				}
			})

		case jsonparser.Object:
			unpackJSON(valueRaw, entry, prefixStr, false)
			return nil

		case jsonparser.String:
			val.string = strings.TrimSpace(valueStr)
			if val.string == "" {
				return nil
			}

		case jsonparser.Number:
			val.float64, _ = jsonparser.ParseFloat(valueRaw)

		case jsonparser.Boolean:
			val.string = valueStr

		default:
			// fmt.Println("unpackJSON: unknown type", prefixStr, valueStr, valueType)
			return nil
		}

		// If top level json
		if topLevel || entry.Level == "" {
			if strings.EqualFold(keyStr, "message") || strings.EqualFold(keyStr, "msg") {
				entry.Message = val.string
				return nil
			} else if strings.EqualFold(keyStr, "level") {
				entry.Level = string(levelFromString(val.string))
				return nil
			}
		}

		// Add tag
		keyStr = conv.String(keyStr)

		if prefix != "" {
			keyStr = prefixStr
		}

		// Add tag
		if val.string != "" {
			if entry.TagsString == nil {
				entry.TagsString = map[string]string{}
			}
			entry.TagsString[keyStr] = val.string
		} else {
			if entry.TagsNumber == nil {
				entry.TagsNumber = map[string]float64{}
			}
			entry.TagsNumber[keyStr] = val.float64
		}

		return nil
	})
}
