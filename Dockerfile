################ STEP 1 #################
FROM golang:1.11-alpine as builder

# Copy source
COPY . /src
WORKDIR /src

ARG VERSION=latest

RUN adduser -D -g '' container
RUN apk update && apk add git && apk add ca-certificates
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -X main.version=${VERSION}" -o /bin/node

################ STEP 2 #################
FROM scratch as main

# Copy binary from first step
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /bin/node /bin/node
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER container
CMD [ "/bin/node" ]