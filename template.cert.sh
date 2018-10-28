echo "Processing {{.Domain}}"
certbot certonly --agree-tos --non-interactive --webroot -w /webroot --email {{ .Admin }} -d {{.Domain}} {{ range $alias := .Aliases }} -d {{$alias}} {{end}}
if [ -d /etc/letsencrypt/live/{{.Domain}}/ ]; then
    cd /etc/letsencrypt/live/{{.Domain}}/
    if [ ! -f dhparam.pem ]; then
        echo "DHPARAM"
        openssl dhparam -dsaparam -out dhparam.pem 4096
    fi
    if [ ! -f {{.Domain}}.client.key ]; then
        echo "Client-CA"
        openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:4096 -keyout {{.Domain}}.client.key -out {{.Domain}}.client.crt -subj "/O={{.Domain}}/CN={{.Domain}}/"
    fi
    if [ ! -d /etc/letsencrypt/live/{{.Domain}}/clients ]; then
        echo "Clients"
        mkdir /etc/letsencrypt/live/{{.Domain}}/clients
    fi
    cd /etc/letsencrypt/live/{{.Domain}}/clients
    {{ with $root := .}}{{ range $client := .Clients }}
    if [ ! -f "{{$client.CommonName}}.p12" ]; then
        echo "Client {{$client.CommonName}}"
        openssl req -out "{{$client.CommonName}}.csr" -new -newkey rsa:2048 -nodes -keyout "{{$client.CommonName}}.key" -subj "/O={{$root.Domain}}/CN={{$client.CommonName}}/"
        openssl x509 -req -days 365 -in "{{$client.CommonName}}.csr" -CA ../{{$root.Domain}}.client.crt -CAkey ../{{$root.Domain}}.client.key -set_serial 01 -out "{{$client.CommonName}}.crt"
        openssl pkcs12 -export -out "{{$client.CommonName}}.p12" -inkey "{{$client.CommonName}}.key" -in "{{$client.CommonName}}.crt" -certfile ../{{$root.Domain}}.client.crt -passout pass:"{{$client.CommonName}}"
    fi
    {{ end }}{{ end }}
fi

