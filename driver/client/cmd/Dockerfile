FROM golang:1.19

WORKDIR /mafia_driver_client

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cd client/cmd && go install
RUN go build -o mafia_driver_client client/cmd/main.go

CMD ["./mafia_driver_client"]
