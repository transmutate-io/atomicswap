FROM golang:1.15.0-alpine3.12 AS build

RUN mkdir -p /src/atomicswap

COPY . /src/atomicswap/

RUN cd /src/atomicswap && go build github.com/transmutate-io/atomicswap/cmd/swapcli


FROM alpine:3.12.0

COPY --from=build /src/atomicswap/swapcli /

RUN mkdir /data

VOLUME [ "/data" ]

ENTRYPOINT [ "/swapcli", "-D", "/data" ]
