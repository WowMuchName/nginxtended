# Combine nginx' and certbots' Dockerfiles
cat "Dockerfile.prefix" > Dockerfile
curl https://raw.githubusercontent.com/certbot/certbot/master/Dockerfile | grep -e "FROM" >> Dockerfile
cat "Dockerfile.middle" >> Dockerfile
echo "# Certbots filtered Dockerfile" >> Dockerfile
curl https://raw.githubusercontent.com/certbot/certbot/master/Dockerfile | grep -v -e "EXPOSE" -e "ENTRYPOINT" -e "FROM" -e "LABEL" -e "VOLUME" -e "CMD" -e "COPY" -e "WORKDIR" >> Dockerfile
echo "# Nginx's filtered Dockerfile" >> Dockerfile
curl https://raw.githubusercontent.com/nginxinc/docker-nginx/master/stable/alpine/Dockerfile | grep -v -e "EXPOSE" -e "ENTRYPOINT" -e "FROM" -e "LABEL" -e "CMD" -e "COPY" -e "VOLUME" -e "WORKDIR" >> Dockerfile
echo "# End of Nginx's Dockerfile" >> Dockerfile
cat "Dockerfile.suffix" >> Dockerfile
docker build -t "wowmuchname/nginxtended" .
