FROM golang:1.25.1-trixie AS build

WORKDIR /app

COPY go.mod ./
#COPY go.sum ./

RUN go mod download && go mod verify

COPY /src .

RUN CGO_ENABLED=0 go build -o sender

FROM scratch

COPY --from=build /app/sender /sender

ENTRYPOINT ["/sender"]
