FROM golang:1.19

WORKDIR /mafia_stat_manager

COPY stat_manager/go.mod stat_manager/go.sum ./
RUN go mod download

COPY . .

RUN cd stat_manager/cmd && go install
RUN go build -o mafia_stat_manager stat_manager/cmd/main.go

ENV CGO_ENABLED=1

CMD ["./mafia_stat_manager"]
