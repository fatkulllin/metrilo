.ONESHELL:

test-iter3:
	rm -f ./cmd/agent/agent ./cmd/server/server; go build -o ./cmd/agent/agent ./cmd/agent/main.go; go build -o ./cmd/server/server ./cmd/server/main.go;
	./metricstest -test.v -test.run="^TestIteration3[AB]*$$" -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

run-test:
	docker run -v "$$(pwd):/app" gotests task unit-tests

# pprof:
# go tool pprof -proto -seconds=30 http://localhost:8080/debug/pprof/heap > profiles/base.pprof
# go tool pprof profiles/base.pprof

private.pem:
	openssl genrsa -out private.pem 2048

public.pem: private.pem
	openssl rsa -in private.pem -pubout -out public.pem

genkeys: private.pem public.pem
