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

COPY . /go/src/swirl/swirlservice
WORKDIR /go/src/swirl/swirlservice
ARG BUILD_TAGS="netgo osusergo"

RUN go build -o swirly-swirlservice -ldflags "-w" *.go
RUN cp swirly-swirlservice /usr/bin/swirly-swirlservice

FROM alpine
RUN echo "http://dl-cdn.alpinelinux.org/alpine/latest-stable/community" >> /etc/apk/repositories
RUN apk update
RUN apk add docker

COPY --from=builder /usr/bin/swirly-swirlservice /usr/bin/swirly-swirlservice
COPY ./defaultconfig.json /usr/bin/defaultconfig.json
COPY ./deployment.json /usr/bin/deployment.json

WORKDIR /usr/bin
ENTRYPOINT [ "./swirly-swirlservice" ] 
