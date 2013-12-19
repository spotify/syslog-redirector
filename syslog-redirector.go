package main

import (
	"bytes"
	"errors"
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"syscall"
)

type Syslogger struct {
	logger *log.Logger
	stream string
	buffer *bytes.Buffer
}

func (s *Syslogger) Write(p []byte) (n int, err error) {
	for b := range p {
		s.buffer.WriteByte(p[b])
		if p[b] == 10 { // newline
			msg := string(s.buffer.Bytes())
			s.logger.Print(msg)
			s.buffer = bytes.NewBuffer([]byte{})
		}
	}
	return len(p), nil
}

func (s *Syslogger) Close() error {
	return nil
}

func NewSysLogger(stream, hostPort, prefix string) (*Syslogger, error) {
	var priority syslog.Priority
	if stream == "stderr" {
		priority = syslog.LOG_ERR | syslog.LOG_LOCAL0
	} else if stream == "stdout" {
		priority = syslog.LOG_INFO | syslog.LOG_LOCAL0
	} else {
		return nil, errors.New("cannot create syslogger for stream " + stream)
	}
	logFlags := 0

	s, err := syslog.Dial("tcp", hostPort, priority, prefix)
	if err != nil {
		return nil, err
	}

	logger := log.New(s, "", logFlags)
	return &Syslogger{logger, stream, bytes.NewBuffer([]byte{})}, nil
}

func usage() {
	log.Printf("usage: %s -h syslog_host:port -n name -- executable [arg ...]", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	log.SetFlags(0)

	flHostPort := flag.String("h", "", "Host port of where to connect to the syslog daemon")
	flLogName := flag.String("n", "", "Name to log as")
	flag.Parse()

	if *flHostPort == "" {
		log.Printf("Must set the syslog host:port argument")
		usage()
	}

	if *flLogName == "" {
		log.Printf("Must set the syslog log name argument")
		usage()
	}

	//Example ./syslog-redirector -h 10.0.3.1:6514 -n test-ls-thingy -- \
	//            /bin/bash -c 'while true; do date; echo $SHELL; sleep 1; done'
	if len(os.Args) < 4 {
		log.Printf("at least 3 arguments required")
		usage()
	}
	hostPort := *flHostPort
	name := *flLogName

	if len(flag.Args()) == 0 {
		log.Printf("must supply a command")
		usage()
	}

	cmdName := flag.Args()[0]
	cmdArgs := flag.Args()[1:]

	var err error

	path, err := exec.LookPath(cmdName)
	if err != nil {
		log.Printf("Unable to locate %v", cmdName)
		os.Exit(127)
	}

	cmd := exec.Command(path, cmdArgs...)

	// TODO (dano): tolerate syslog downtime by reconnecting

	cmd.Stdout, err = NewSysLogger("stdout", hostPort, name)
	if err != nil {
		log.Printf("error creating syslog writer for stdout: %v", err)
	}

	cmd.Stderr, err = NewSysLogger("stderr", hostPort, name)
	if err != nil {
		log.Printf("error creating syslog writer for stderr: %v", err)
	}

	err = cmd.Run()
	if err != nil {
		if msg, ok := err.(*exec.ExitError); ok {
			os.Exit(msg.Sys().(syscall.WaitStatus).ExitStatus())
		} else {
			log.Printf("error running command: %v", err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}
