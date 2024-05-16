FROM golang:1.22 AS build
WORKDIR /go/src
COPY go ./go
COPY main.go .


ENV CGO_ENABLED=0
RUN go get -d -v ./...

RUN go build -a -installsuffix cgo -o swagger .

FROM scratch AS runtime
COPY --from=build /go/src/swagger ./
COPY .env .
COPY data.json .

ARG SERVER_PORT=8080

EXPOSE ${SERVER_PORT}/tcp
ENTRYPOINT ["./swagger"]
