FROM alpine:edge

RUN mkdir -p /build/server/src /app /web
RUN apk add --no-cache go nodejs

ADD Server/src /build/server/src

ENV GOPATH=/build/server
ENV GOBIN=/app

RUN apk add --no-cache -t tmp git musl-dev
RUN cd /build/server \
	&& go get github.com/lib/pq \
	&& go install src/kvj/autohome/server_run.go \
	&& rm -r /build \
	&& apk del -q tmp

RUN mkdir -p /web
ADD Web /web

RUN cd /web \
	&& npm install --development \
	&& node_modules/.bin/gulp dist \
	&& rm -r node_modules static/less static/coffee

EXPOSE 9100

CMD ["/app/server_run"]

