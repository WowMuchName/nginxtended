FROM golang:alpine as gobuilder
ADD . /go
RUN go build -o configurator
FROM nginx:alpine
RUN apk add --no-cache openssl certbot
ADD nginx.conf global.conf /etc/nginx/
COPY --from=gobuilder /go/configurator /go/template.https.conf /go/template.stream.conf /go/template.cert.sh /var/lib/configurator/
RUN mkdir /webroot \
 && mkdir /etc/nginx/stream-conf.d
WORKDIR /var/lib/configurator
ENTRYPOINT ["/var/lib/configurator/configurator"]
