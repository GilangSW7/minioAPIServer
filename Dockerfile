FROM golang:1.19-alpine as build

##Build
WORKDIR /app

COPY . .

RUN go mod download


RUN go build -o /docker-minio-api

FROM alpine
WORKDIR /app
COPY --from=build /docker-minio-api /docker-minio-api

EXPOSE 8090

CMD ["/docker-minio-api"]
