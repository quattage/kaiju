package console

import (
	"strconv"
	"strings"
	"unicode"

	"kaijuengine.com/engine/systems/logging"
)

type logLevel uint8

const (
	LogLevelInfo logLevel = iota
	LogLevelWarn
	LogLevelError
	LogLevelDebug
)

type logEntry struct {
	Level   logLevel
	Time    string
	Message string
	Trace   []string
}

func (c *Console) MessageAt(index int) *logEntry {
	return &c.messages[index]
}

func (c *Console) MessageCount() int {
	return len(c.messages)
}

func (c *Console) AddOnMessageSent(oms func(msg string)) logging.EventId {
	return c.evt.Add(oms)
}

func (c *Console) RemoveOnMessageSent(evtid logging.EventId) {
	c.evt.Remove(evtid)
}

func (c *Console) SetSize(size uint16) {
	c.scrollback = size
	if len(c.messages) > int(c.scrollback) {
		c.messages = c.messages[len(c.messages)-int(c.scrollback):]
	}
}

func (c *Console) ParseNewLogMessage(lvl logLevel, input string) {

	if len(input) <= 0 {
		return
	}

	var fields []string
	var working strings.Builder
	var quoted bool
	var escaped bool

	c.mutex.Lock()
	for _, r := range input {
		if escaped {
			working.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quoted {
			working.WriteRune(r)
			escaped = true
			continue
		}
		if r == '"' {
			quoted = !quoted
			working.WriteRune(r)
			continue
		}
		if unicode.IsSpace(r) && !quoted {
			if working.Len() > 0 {
				fields = append(fields, working.String())
				working.Reset()
			}
			continue
		}
		working.WriteRune(r)
	}
	if working.Len() > 0 {
		fields = append(fields, working.String())
	}

	entry := &logEntry{}
	var varargs strings.Builder

	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		if len(value) >= 2 && value[0] == '"' {
			value, _ = strconv.Unquote(value)
		}
		switch key {
		case "time":
			entry.Time = value
			startidx := strings.Index(entry.Time, "T")
			endidx := strings.LastIndex(entry.Time, "-")
			if startidx > 0 && endidx > 0 {
				entry.Time = entry.Time[startidx+1 : endidx]
			}
		case "level":
			if strings.HasPrefix(value, "D") {
				entry.Level = LogLevelDebug
			} else {
				entry.Level = lvl
			}
		case "msg":
			entry.Message = value
		default:
			if varargs.Len() > 0 {
				varargs.WriteString(", ")
			}
			varargs.WriteString(strings.Replace(field, "=", ": ", 1))
		}
	}
	entry.Message = strings.TrimSpace(entry.Message)
	if varargs.Len() > 0 {
		entry.Message += " [" + strings.TrimSpace(varargs.String()) + "]"
	}
	c.messages = append(c.messages, *entry)
	if len(c.messages) > int(c.scrollback) {
		c.messages = c.messages[len(c.messages)-int(c.scrollback):]
	}
	c.mutex.Unlock()
	c.evt.Execute(entry.Message)
}
