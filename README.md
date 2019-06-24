# Nginxtended

A docker-image containing Nginx and Certbot, based on alpine-linux, wrapped into a convenient package optimized for SSL-termination use-cases.

## Features
* ~90MB small image based on alpine
* Full versions of Nginx and Certbot
* Drop-in replacement for the official nginx image
* Modified Nginx root-configuration (/etc/nginx/nginx.conf) that is tweaked for use as a reverse proxy with docker and Certbot
* Modified default settings for nginx that should give you an A+ on SSL-labs out of the box
* Ability to configure both certbot and nginx using one, json-based configuration file
* Un-restricted access to nginx's and certbot's normal (low-level) configuration files
* Automatically obtains and renews the certificates you need from letsencrypt
* Add, remove, modify V-Hosts without service interruptions

## Usage
Create this *docker-compose.yml*:
```yml
version: "3.7"
services:
  nginxtended:
    image: "wowmuchname/nginxtended"
    volumes:
      - certs:/etc/letsencrypt
      - backends:/backends
    ports:
      - 80:80
      - 443:443
    networks:
      - proxy-net
    container_name: nginxtended      
volumes:
  certs:
    external: true
  backends:
    external: true
networks:
  proxy-net:
    external: true
```
Create the network/volumes that persist between restarts
```sh
docker volume create certs
docker volume create backends
docker network create proxy-net
```
Start the container
```sh
docker-compose up -d
```
Start the container you want to provide SSL-Termination for. For instance:
```sh
docker run -d --network proxy-net --name test nginx
```

Create your V-Host(s) */var/lib/docker/volumes/backends/_data/test.json*
```json
{
    "domain": "test.example.com",
    "url": "http:/test:80"
}
```

Reload nginxtended
```
docker exec -it nginxtended configurator reload
```
This will obtain a certificate for *test.example.com* from letsencrypt and make the container with the name *test* available via
https://test.example.com

Nginxtended will try to renew the certificate (which is valid for 90 days) once a week. The renewal happens when the cert is valid for
less than 30 days.

**Note:** You only need to reload when you add, remove or change V-Hosts. Always use the command as described above. This command relies on
nginx's ability to reload config files **WITHOUT** restarting. This means you can change stuff without downtime.

## Backend-Config (aka VHost-Config)
The config files reside at */backends* in the container. When nginxtended starts or reloads those are parsed and config-files for nginx
and certbot are created from them. This process also deletes derived config files if the backend-config was deleted. Note that you can still
add nginx config-files directly to /etc/nginx/conf.d. As long as they do not start with the prefix *derived_* the configurator will never
modify or remove them.

example.json

```json
{
    "version": "1.0",
    "domain": "test.example.com",
    "url": "http:/test:80",
    "admin": "admin@example.com",
    "protocol": "https",
    "port": 443,
    "aliases": ["test2.example.com"],
    "clients": [{
        "CommonName": "Jon Doe"
    }],
    "keyauth": false
}
```

**NOTE** The configurator is written in GO and uses its strict JSON parser. You cannot have unquoted property-names, trailing colons or comments
in your JSON file.

|Property|Default|Description|
|--------|-------|-----------|
|version | 1.0   | Must be 1.0 or omitted |
|domain  | required | Fully qualified domain-name the vhost will be reachable at |
|url | required | Url nginx will proxy to (value for proxy_pass) |
| protocol | https | Can be https or tls. The later will create a TCP proxy |
| port | 443 | Port to to listen on |
| aliases | empty | Nginx will permanently redirect call to this domains to the main domain |
| clients | empty | Client-Keys will be created for those names if they do not exist |
| keyauth | true if clients is not empty | Nginx will require client-keys for this V-Host |

## Beyond Cerbot

### Client-Certificates
Apart from letting certbot obtain certificates for Https/TLS-Termination, nginxtended creates a server-key and certificate for use with client authentication (if they do not already exist). They are put into *certs-volume*/live/*fqdn*, that is the same place cerbot links the certificates and keys used for SSL.

### Automatic client-certificate creation
If you specify clients in your config, a subdirectory is created containing a private-key, a certificate and a p12 Cryptographic storage for each client of the V-Host (they are not shared among hosts). The p12 file uses the client's name as specified in the config for a password.

### Manual client-certificate creation (recommended)
Client-Certificates can also be generated manually. Nginxtended comes with a handy script called make-user-cert.sh. This script is copied into the *certs* volume every time nginxtended starts. Provided you have openssl installed on the host and you are using a hostmount as your *certs* volume, you can create a client-certificate directly on the host.
```
/var/lib/docker/volume/*certs*/_data/make-user-cert.sh *UserName* /var/lib/docker/volume/*certs*/_data/live/*domain*
```
The command will ask you for a password and create a encrypted p12 file in the current working directory.

### Perfect Forward Secrecy
Nginxtended creates individual ssl_dhparam files for each V-Host if they do not already exist. They are put into /certs/live/DOMAIN.

## Low level usage
Note that while you can still use this image as a replacement for nginx's image, the automatic acquisition of certificates, perfect forward secrecy
and some other features rely on the fact that you use nginxtended's config files. If you configure nginx directly you need to take care of
those things yourself.

## Scope of this project
This project is meant for cases where you do not want to use kubernetes or a similar technology but still have a single host that hosts
many vhosts.

This means the project does not aim to cover cases where v-hosts are distributed among multiple machines. While you can still use it
that way, you might be better of using the options your container-service provides.

## License
All the files in the repository are licensed under MIT.

The docker-image however contains code of third-parties.
