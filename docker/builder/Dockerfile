FROM golang:latest

RUN apt-get update && apt-get upgrade -y
RUN apt-get install python3 -y && \
    apt-get install python3-pip \

WORKDIR /app

COPY . .

RUN go build -o app cmd/builder/main.go

# accepts 3 args - see usage in builder.go
ENTRYPOINT [ "./app" ]