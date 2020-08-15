FROM golang:1.14.3-alpine3.11 AS build

RUN mkdir -p /src/atomicswap

COPY . /src/atomicswap/

RUN cd /src/atomicswap && go build github.com/transmutate-io/atomicswap/cmd/swapcli


FROM alpine:3.12.0

COPY --from=build /src/atomicswap/swapcli /

RUN mkdir /data

VOLUME [ "/data" ]

ENTRYPOINT [ "/swapcli", "-D", "/data" ]
