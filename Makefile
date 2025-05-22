.ONESHELL:

test-iter3:
	rm -f ./cmd/agent/agent ./cmd/server/server; go build -o ./cmd/agent/agent ./cmd/agent/main.go; go build -o ./cmd/server/server ./cmd/server/main.go;
	./metricstest -test.v -test.run="^TestIteration3[AB]*$$" -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server
