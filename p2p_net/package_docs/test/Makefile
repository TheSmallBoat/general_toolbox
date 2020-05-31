cd ./p2p_network/package_docs/test

test-coverage:
	go test -v -coverprofile=coverage.txt -covermode=atomic -timeout=5m -race ../../...

test:
	go test -v -timeout=5m -race ../../...

test bench:
    cd ./p2p_network/core_module
	go test -timeout=5m -bench=.