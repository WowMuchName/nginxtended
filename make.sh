# Compile the go-code
docker run --rm -it -v "${PWD}":/go/ golang go build -o configurator

# Checkout Certbot from GIT and remove unneeded files
git -C certbot pull || git clone https://github.com/certbot/certbot.git certbot
mkdir -p certbot-slim
cp certbot/README.rst certbot-slim/
cp certbot/setup.py certbot-slim/
cp -r certbot/acme certbot-slim/
cp -r certbot/certbot certbot-slim/
rm -r certbot-slim/acme/acme/testdata
rm -r certbot-slim/acme/docs
rm -r certbot-slim/acme/examples
rm -r certbot-slim/certbot/tests

# Combine nginx' and certbots' Dockerfiles
cat "Dockerfile.prefix" > Dockerfile
curl https://raw.githubusercontent.com/certbot/certbot/master/Dockerfile | grep -v -e "EXPOSE" -e "ENTRYPOINT" -e "LABEL" -e "FROM" -e "VOLUME" -e "CMD" -e "COPY" -e "WORKDIR" >> Dockerfile
curl https://raw.githubusercontent.com/nginxinc/docker-nginx/master/stable/alpine/Dockerfile | grep -v -e "EXPOSE" -e "ENTRYPOINT" -e "FROM" -e "LABEL" -e "CMD" -e "COPY" -e "VOLUME" -e "WORKDIR" >> Dockerfile
cat "Dockerfile.suffix" >> Dockerfile
docker build -t "wowmuchname/nginxtended" .
