# #build stage
# FROM golang:1.16.2-alpine3.13 AS builder
# RUN apk add --no-cache git
# WORKDIR /go/src/app
# COPY . .
# RUN go get -d -v ./...
# RUN go install -v ./...

# #final stage
# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
# COPY --from=builder ./go/bin/app ./app
# COPY ./web/dist ./web/dist
# ENTRYPOINT ./app
# LABEL Name=pingerng Version=0.0.1
# EXPOSE 6060

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY ./pingerng ./
COPY ./tabPages ./tabPages
COPY ./web/dist ./web/dist
CMD ["./pingerng"]
LABEL Name=pingerng Version=0.0.1
EXPOSE 6060
