FROM golang:1.21

# RUN apt-get install git jq curl -y
RUN go version

COPY . .

# RUN cd /app && go mod download

RUN GOOS=linux GOARCH=amd64 go build -o HugoToMedium main.go

RUN chmod +x entrypoint.sh

RUN chmod +x HugoToMedium

ENTRYPOINT [ "entrypoint.sh" ]
