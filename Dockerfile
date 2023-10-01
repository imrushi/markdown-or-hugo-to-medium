FROM golang:1.21.1-alpine3.18

# RUN apt-get install git jq curl -y
RUN apk add git

COPY . /home/src

WORKDIR /home/src

RUN GOOS=linux GOARCH=amd64 go build -o HugoToMedium main.go

RUN chmod +x entrypoint.sh

RUN chmod +x HugoToMedium

ENTRYPOINT [ "entrypoint.sh" ]
