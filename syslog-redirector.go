package main

import (
	"errors"
	"fmt"
	"log"
	"log/syslog"
"os"
"os/exec"
)

type Syslogger struct {
	logger *log.Logger
	stream string
}

func (s *Syslogger) Write(p []byte) (n int, err error) {
	x := string(p)
	s.logger.Println(x)
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
	return &Syslogger{logger, stream}, nil
}

func main() {
	//args are: syslog_host:port name command to run
	//example ./syslog-redirector 10.0.3.1:6514 test-ls-thingy \
        //            /bin/bash -c 'while true; do date; echo $SHELL; sleep 1; done'

	hostPort := os.Args[1]
	name := os.Args[2]

	cmdArgs := os.Args[4:]
	cmd := exec.Command(os.Args[3], cmdArgs...)
	var err error
	cmd.Stdout, err = NewSysLogger("stdout",  hostPort, name)
	if err != nil {
		fmt.Errorf("error creating syslog writer: " + err.Error())
	}
	cmd.Stderr, err = NewSysLogger("stderr",  hostPort, name)
	err = cmd.Run()
	if err != nil {
		fmt.Errorf("error running command: " + err.Error())
	}
}










