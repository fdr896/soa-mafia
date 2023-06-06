FROM golang:1.19

WORKDIR /mafia_driver_server

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd server/cmd && go install
RUN go build -o mafia_driver_server server/cmd/main.go

ENV CGO_ENABLED=1

CMD ["./mafia_driver_server"]
