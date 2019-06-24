FROM golang:alpine as gobuilder
ADD . /go
RUN go build -o configurator
FROM nginx:alpine
RUN apk add --no-cache openssl certbot
ADD nginx.conf global.conf /etc/nginx/
COPY --from=gobuilder /go/configurator /go/template.https.conf /go/template.stream.conf /go/template.cert.sh /go/make-user-cert.sh /var/lib/configurator/
RUN mkdir /webroot \
 && mkdir /etc/nginx/stream-conf.d \
 && chmod 0755 /var/lib/configurator/make-user-cert.sh \
 && ln -s /var/lib/configurator/configurator /usr/bin/configurator
WORKDIR /var/lib/configurator
ENTRYPOINT ["/var/lib/configurator/configurator"]
