test:
	go test ./internal/core/distributions/ -v
testSim:
	go test ./internal/core/distributions/sim/ -v
coverage:
	go test ./internal/core/distributions/ -cover
  
