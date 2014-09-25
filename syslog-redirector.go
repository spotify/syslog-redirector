//  Copyright (c) 2014 Spotify AB.
// 
//  Licensed to the Apache Software Foundation (ASF) under one
//  or more contributor license agreements.  See the NOTICE file
//  distributed with this work for additional information
//  regarding copyright ownership.  The ASF licenses this file
//  to you under the Apache License, Version 2.0 (the
//  "License"); you may not use this file except in compliance
//  with the License.  You may obtain a copy of the License at
// 
//    http://www.apache.org/licenses/LICENSE-2.0
// 
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an
//  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
//  KIND, either express or implied.  See the License for the
//  specific language governing permissions and limitations
//  under the License.

package main

import (
	"bytes"
	"errors"
	"flag"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Syslogger struct {
	logger   *Writer
	stream   string
	buffer   *bytes.Buffer
	hostPort string
	priority Priority
	prefix   string
	logFlags int
        protocol string
}

func (s *Syslogger) Write(p []byte) (n int, err error) {
	if s.logger == nil {
		sl, err := Dial(s.protocol, s.hostPort, s.priority, s.prefix)
		if err != nil {
			// while syslog is down, dump the output
			return len(p), nil
		}
		s.logger = sl
	}
	for b := range p {
		s.buffer.WriteByte(p[b])
		if p[b] == 10 { // newline
			_, err := s.logger.Write(s.buffer.Bytes())
			if err != nil {
				s.logger = nil
			}
			s.buffer = bytes.NewBuffer([]byte{})
		}
	}
	return len(p), nil
}

func (s *Syslogger) Close() error {
	return nil
}

func NewSysLogger(stream, hostPort, prefix, protocol string) (*Syslogger, error) {
	var priority Priority
	if stream == "stderr" {
		priority = LOG_ERR | LOG_LOCAL0
	} else if stream == "stdout" {
		priority = LOG_INFO | LOG_LOCAL0
	} else {
		return nil, errors.New("cannot create syslogger for stream " + stream)
	}
	logFlags := 0

	return &Syslogger{nil, stream, bytes.NewBuffer([]byte{}), hostPort, priority, prefix, logFlags, protocol}, nil
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
	flProtocol := flag.Bool("t", false, "use TCP instead of UDP (the default) for syslog communication")
	flag.Parse()

	if *flHostPort == "" {
		log.Printf("Must set the syslog host:port argument")
		usage()
	}

	if *flLogName == "" {
		log.Printf("Must set the syslog log name argument")
		usage()
	}

	protocol := "udp"
	if *flProtocol {
		protocol = "tcp"
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

	cmd.Stdout, err = NewSysLogger("stdout", hostPort, name, protocol)
	if err != nil {
		log.Printf("error creating syslog writer for stdout: %v", err)
	}

	cmd.Stderr, err = NewSysLogger("stderr", hostPort, name, protocol)
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
