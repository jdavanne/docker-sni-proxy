FROM golang:alpine AS build
RUN apk add --no-cache make git
WORKDIR /app/src

COPY . ./
RUN go build -o dummy-tlsserver -ldflags "-s -w "

EXPOSE 443

CMD ["./dummy-tlsserver"]
