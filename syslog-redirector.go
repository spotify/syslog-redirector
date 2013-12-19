package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"os/exec"
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
	fmt.Errorf("usage: %s -h syslog_host:port -n name -- executable [arg ...]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flHostPort := flag.String("h", "", "Host port of where to connect to the syslog daemon")
	flLogName := flag.String("n", "", "Name to log as")
	flag.Parse()

	if *flHostPort == "" {
		fmt.Println("Must set the syslog host:port argument")
		usage()
	}

	if *flLogName == "" {
		fmt.Println("Must set the syslog log name argument")
		usage()
	}

	//Example ./syslog-redirector -h 10.0.3.1:6514 -n test-ls-thingy -- \
	//            /bin/bash -c 'while true; do date; echo $SHELL; sleep 1; done'
	if len(os.Args) < 4 {
		fmt.Printf("at least 3 arguments required\n")
		usage()
	}
	hostPort := *flHostPort
	name := *flLogName

	if len(flag.Args()) == 0 {
		fmt.Printf("must supply a command")
		usage()
	}

	cmdArgs := flag.Args()[1:]
	cmd := exec.Command(flag.Args()[0], cmdArgs...)

	var err error

	// TODO (dano): tolerate syslog downtime by reconnecting

	cmd.Stdout, err = NewSysLogger("stdout", hostPort, name)
	if err != nil {
		fmt.Errorf("error creating syslog writer for stdout: " + err.Error())
	}

	cmd.Stderr, err = NewSysLogger("stderr", hostPort, name)
	if err != nil {
		fmt.Errorf("error creating syslog writer for stderr: " + err.Error())
	}

	err = cmd.Run()
	if err != nil {
		fmt.Errorf("error running command: " + err.Error())
	}
}
