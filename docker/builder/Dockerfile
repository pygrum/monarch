FROM golang:latest

RUN apt-get update && apt-get upgrade -y
RUN apt-get install python3 -y && \
    apt-get install python3-pip -y

# install monarch SDK
RUN rm /usr/lib/python3.11/EXTERNALLY-MANAGED
RUN python3 -m pip install -i https://test.pypi.org/simple/ monarch-c2-sdk==0.0.3

WORKDIR /go/delivery

COPY . .

RUN go build -o app cmd/builder/main.go

# accepts 3 args - see usage in builder.go
ENTRYPOINT [ "/go/delivery/app" ]