FROM golang:1.3.3

ADD src/ /go/src
RUN CGO_ENABLED=0 go build -o /syslog-redirector -tags netgo -a /go/src/*.go

# Verify that we built a completely static binary
RUN ldd /syslog-redirector; test $? -eq 1
