FROM golang:latest

RUN apt update -y
RUN apt install -y curl
RUN curl https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz \
    | tar xvz -C /tmp/ && mv /tmp/docker/docker /usr/bin/docker

WORKDIR /go/delivery

COPY . .

RUN echo 'export PATH=$PATH:$HOME/.local/bin' >> $HOME/.profile