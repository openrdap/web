package controllers

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/cosiner/argv"
	"github.com/openrdap/rdap"
	"github.com/revel/revel"
	"golang.org/x/net/websocket"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) API() revel.Result {
	return c.Render()
}

func (c App) Docs() revel.Result {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	opts := rdap.CLIOptions{
		Sandbox: true,
	}

	rdap.RunCLI([]string{"--help"}, &stdout, &stderr, opts)

	c.ViewArgs["HelpText"] = stdout.String()

	return c.Render()
}

func (c App) Demo() revel.Result {
	cmd := c.Params.Query.Get("cmd")
	if cmd == "" {
		cmd = "rdap -v -e example.com"
	}

	c.ViewArgs["cmd"] = cmd
	return c.Render()
}

type message struct {
	Text string
}

func (c App) RDAPWebsocket(ws *websocket.Conn) revel.Result {
	errorAdvice := "Invalid command. Try 'rdap example.cz'"

	cmd := c.Params.Query.Get("cmd")
	if !strings.HasPrefix(cmd, "rdap") {
		websocket.JSON.Send(ws, message{Text: errorAdvice})
		return nil
	}
	cmd = strings.TrimPrefix(cmd, "rdap")
	cmd = strings.TrimPrefix(cmd, " ")

	env := map[string]string{}
	quoteParser := func([]rune, map[string]string) ([]rune, error) {
		return []rune{}, nil
	}

	cmds, err := argv.Argv([]rune(cmd), env, quoteParser)
	if err != nil || len(cmds) != 1 {
		websocket.JSON.Send(ws, message{Text: errorAdvice})
		return nil
	}

	opts := rdap.CLIOptions{
		Sandbox: true,
	}

	stdout := websocketWriter{Websocket: ws}
	stderr := websocketWriter{Websocket: ws}

	rdap.RunCLI(cmds[0], &stdout, &stderr, opts)

	return nil
}

type websocketWriter struct {
	Websocket *websocket.Conn
}

func (w *websocketWriter) Write(p []byte) (int, error) {
	err := websocket.JSON.Send(w.Websocket, &message{
		Text: html.EscapeString(fmt.Sprintf("%s", p)),
	})

	if err != nil {
		return 0, err
	}

	return len(p), nil
}
