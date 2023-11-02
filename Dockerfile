FROM golang:1.21.1-alpine3.18

# RUN apt-get install git jq curl -y
RUN apk add git

COPY . /home/src

WORKDIR /home/src

RUN GOOS=linux GOARCH=amd64 go build -o md-hugo-to-medium main.go

RUN chmod +x md-hugo-to-medium

ENTRYPOINT [ "/home/src/md-hugo-to-medium" ]
