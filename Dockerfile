FROM golang:1.15-alpine
WORKDIR /src

COPY go.sum go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /bin/app .

FROM alpine
WORKDIR /usr/src

COPY --from=0 /bin/app ./app

ENTRYPOINT ["/usr/src/app"]
