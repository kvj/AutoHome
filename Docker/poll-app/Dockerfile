FROM armhf/alpine:edge

RUN mkdir -p /build/src /app
RUN apk add --no-cache go

ADD Server/src /build/src

ENV GOPATH=/build
ENV GOBIN=/app

RUN apk add --no-cache -t tmp git musl-dev
RUN cd /build \
	&& go get github.com/lib/pq \
	&& go get github.com/tarm/goserial \
	&& go install src/kvj/autohome/poll_run.go \
	&& rm -r /build \
	&& apk del -q tmp

# create table measure (id serial primary key, device int, type int, sensor int, measure int, value float, at timestamp with time zone);
# create table forecast (id serial primary key, device int, type int, sensor int, measure int, value float, at timestamp with time zone);
# create index measure_index0 on measure (device, type, sensor, measure, at).;

CMD ["/app/poll_run"]

