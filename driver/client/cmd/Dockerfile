FROM golang:1.19

WORKDIR /mafia_driver_client

COPY driver/go.mod driver/go.sum ./
RUN go mod download

COPY . .

RUN cd driver/client/cmd && go install
RUN go build -o mafia_driver_client driver/client/cmd/main.go

ENV CGO_ENABLED=1

CMD ["./mafia_driver_client"]
