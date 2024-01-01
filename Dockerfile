FROM golang:alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GIT_COMMIT=empty
RUN go build -a -ldflags "-X main.GIT_COMMIT=$GIT_COMMIT -s -w" \
  -gcflags="all=-trimpath=${PWD}" \
  -asmflags="all=-trimpath=${PWD}" \
  -o start

FROM alpine
WORKDIR /app
COPY --from=builder /app/start .
RUN apk add tzdata

VOLUME [ "/app/data" ]
EXPOSE 8080
ENV GIN_MODE=release

CMD /app/start
