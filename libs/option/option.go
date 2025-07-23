/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package option

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

type Http struct {
	Path         string `long:"http.path" default:"" description:"Path for the HTTP server context" `
	Address      string `long:"http.address" default:"0.0.0.0" description:"Address for the HTTP server listening" `
	Port         int    `long:"http.port" default:"8080" description:"Port for the HTTP server listening" `
	Cors         bool   `long:"http.cors" description:"Support CORS access" `
	Access       bool   `long:"http.access" description:"Access control for HTTP server" `
	RequestLog   bool   `long:"http.requestlog" description:"Log HTTP requests" `
	Trace        bool   `long:"http.trace" description:"Trace HTTP requests" `
	IdleTimeout  int    `long:"http.idle" default:"0" description:"Timeout (in seconds) for idle connection" `
	ReadTimeout  int    `long:"http.read" default:"0" description:"Timeout (in seconds) for reading client request" `
	WriteTimeout int    `long:"http.write" default:"0" description:"Timeout (in seconds) for writing to client request" `
}

// Log logging settings
type Log struct {
	Path   string `long:"log.path" default:"logs" required:"true" description:"Sets the path to log file"`
	Level  string `long:"log.level" default:"info" description:"Sets the log level" choice:"info" choice:"warn" choice:"error" choice:"debug" choice:"trace" `
	Format string `long:"log.format" default:"text" description:"Sets the log format" choice:"text" choice:"json"`
	Rotate string `long:"log.rotate" default:"day" description:"Rotates the log" choice:"day" choice:"hour" `
}

// Options 服务参数选项
type Options struct {
	Application string `long:"application" required:"true" description:"Application name for startup"`
	Profile     string `long:"profile" required:"true" description:" Profile for startup"`
	ConfigFile  string `long:"config" required:"true" description:"Config file for startup"`
	Log         Log    `group:"log"`
	Http        Http   `group:"http"`
	Version     bool   `long:"version" short:"v" description:"Show the program version"`
}

var _parser *flags.Parser

func NewOptions() *Options {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	var opts Options
	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_parser = parser
	return &opts
}

func (m *Options) AddCommand(name string, cmd flags.Commander) {
	_parser.AddCommand(name, "", "", cmd)
}

func (m *Options) Parse() error {
	_, err := _parser.ParseArgs(os.Args[1:])
	if nil == err {
		return nil
	}
	switch err.(type) {
	case *flags.Error:
		flagError := err.(*flags.Error)
		if flagError.Type == flags.ErrHelp {
			_parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		if flagError.Type == flags.ErrRequired && m.Version {
			os.Exit(0)
		}
		os.Stdout.WriteString("Fault: \n" + err.Error() + "\n")
	default:
		log.Fatal("Unknown error: ", err)
	}

	return err
}
