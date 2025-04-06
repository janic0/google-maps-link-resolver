FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git && apk add ca-certificates
RUN apk --update add ca-certificates
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
RUN go get -v
RUN go build -o /go/bin/google-maps-link-resolver

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/google-maps-link-resolver /go/bin/google-maps-link-resolver
ENTRYPOINT ["/go/bin/google-maps-link-resolver"]
