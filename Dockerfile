FROM golang:1.20.5 AS build

COPY . /root
RUN cd /root/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gcsi

FROM centos:centos7.9.2009
RUN yum -y install e2fsprogs
COPY --from=build /root/gcsi /usr/local/bin/gcsi
ENTRYPOINT ["/usr/local/bin/gcsi"]
