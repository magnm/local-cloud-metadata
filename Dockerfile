FROM golang:1.21 AS builder

WORKDIR /src
COPY ./ ./
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -o lcm

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/lcm /
ENTRYPOINT [ "/lcm" ]