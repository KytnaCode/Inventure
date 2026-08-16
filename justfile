certFile := "server.crt"
keyFile := "server.key"

init: gen-sll

gen-ssl:
  openssl req -x509 -noenc \
    -subj "/C=US/ST=Real-State/L=Real-City/O=Dev/CN=localhost" \
    -newkey rsa -pkeyopt rsa_keygen_bits:4096 -keyout {{keyFile}} -keyform PEM \
    -out {{certFile}} -outform PEM

fmt:
  golangci-lint fmt ./...

lint:
  golangci-lint run ./...

test TEST:
  go test -run ^Test{{TEST}} ./...

test-all:
  go test ./...

test-all-short:
  go test -short ./...

precommit: fmt lint test-all
