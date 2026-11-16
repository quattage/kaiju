package console

import (
	"log/slog"
	"strings"
	"sync"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/logging"
)

var consoles = map[*engine.Host]*Console{}

type Console struct {
	cmds       map[string]executor
	paths      map[string][]string
	host       weak.Pointer[engine.Host]
	scrollback uint16
	messages   []logEntry
	mutex      sync.Mutex
	evt        logging.Event
	evtids     [3]logging.EventId
}

func For(host *engine.Host) *Console {
	c, ok := consoles[host]
	if !ok {
		c = initialize(host)
		consoles[host] = c
		host.OnClose.Add(func() {
			if console, ok := consoles[host]; ok {
				takedown(console, host)
				delete(consoles, host)
			}
		})
	}
	return c
}

func initialize(host *engine.Host) *Console {
	console := &Console{
		host:  weak.Make(host),
		cmds:  make(map[string]executor),
		paths: make(map[string][]string),
	}
	console.evtids[LogLevelInfo] = host.LogStream.OnInfo.Add(func(msg string) {
		console.ParseNewLogMessage(LogLevelInfo, msg)
	})
	console.evtids[LogLevelWarn] = host.LogStream.OnWarn.Add(func(msg string, trace []string) {
		console.ParseNewLogMessage(LogLevelWarn, msg)
	})
	console.evtids[LogLevelError] = host.LogStream.OnError.Add(func(msg string, trace []string) {
		console.ParseNewLogMessage(LogLevelError, msg)
	})
	return console
}

func takedown(console *Console, host *engine.Host) {
	host.LogStream.OnInfo.Remove(console.evtids[LogLevelInfo])
	host.LogStream.OnWarn.Remove(console.evtids[LogLevelWarn])
	host.LogStream.OnError.Remove(console.evtids[LogLevelError])
}

type executor struct {
	action any
	desc   func() string
}

func (c *Console) Execute(command string) {
	if len(command) <= 0 {
		return
	}
	command = strings.TrimSpace(command)
	tokens := strings.Split(command, " ")
	if len(tokens) <= 0 {
		return
	}
	for _, token := range tokens {
		token = strings.TrimSpace(token)
	}
	var acc strings.Builder
	for x, token := range tokens {
		acc.WriteString(token)
		if len(token) <= 0 {
			continue
		}
		cmd, ok := c.cmds[token]
		if ok {
			cmd.Run(tokens[x+1:])
			return
		}
		_, hasNext := c.paths[token]
		if !hasNext {
			return
		}
		acc.WriteString(" ")
	}
	slog.Error("No such command '" + acc.String() + "'")
}

func (c *Console) Host() *engine.Host {
	return c.host.Value()
}

func (ex *executor) Run(varargs []string) {

}
