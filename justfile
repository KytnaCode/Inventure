certFile := "server.crt"
keyFile := "server.key"

init: gen-ssl copy-conf

gen-ssl:
  openssl req -x509 -noenc \
    -subj "/C=US/ST=Real-State/L=Real-City/O=Dev/CN=localhost" \
    -newkey rsa -pkeyopt rsa_keygen_bits:4096 -keyout {{keyFile}} -keyform PEM \
    -out {{certFile}} -outform PEM

copy-conf:
  cp config.example.yml config.yml

fmt:
  golangci-lint fmt ./...

lint:
  golangci-lint run ./...
  cd frontend && pnpm run lint

test TEST:
  go test -run ^Test{{TEST}} ./...

test-all:
  go test -parallel 4 ./...
  cd frontend && pnpm run test

test-all-short:
  go test -short ./...
  cd frontend && pnpm run test

precommit: fmt lint test-all
