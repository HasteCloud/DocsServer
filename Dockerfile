# There appears to be an issue with 1.11 and Go Modules preventing builds from succeeding
# I just changed to the rc of 1.12 and it seemed to work a treat
FROM golang:1.12-rc-alpine as builder
RUN apk add --no-cache gcc curl wget git

WORKDIR /go/app
ADD . .
WORKDIR src/main/
RUN GO111MODULE=on CGO_ENABLE=0 GOOS=linux go build -o mddocs

FROM alpine
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /go/app/src/main/mddocs /app/mddocs
COPY --from=builder /go/app/assets /app/assets

CMD ["./mddocs", "-port=8080"]