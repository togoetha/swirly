FROM golang:alpine as builder
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN apk add --no-cache \
	ca-certificates \
	--virtual .build-deps \
	git \
	gcc \
	libc-dev \
	libgcc \
	make \
	bash 

COPY . /go/src/swirl/edgeservice
WORKDIR /go/src/swirl/edgeservice
ARG BUILD_TAGS="netgo osusergo"

RUN go build -o swirly-edgeservice -ldflags "-w" *.go
RUN cp swirly-edgeservice /usr/bin/swirly-edgeservice

FROM alpine
RUN echo "http://dl-cdn.alpinelinux.org/alpine/latest-stable/community" >> /etc/apk/repositories
RUN apk update
RUN apk add docker

COPY --from=builder /usr/bin/swirly-edgeservice /usr/bin/swirly-edgeservice
COPY ./defaultconfig.json /usr/bin/defaultconfig.json

WORKDIR /usr/bin
ENTRYPOINT [ "./swirly-edgeservice" ] 
