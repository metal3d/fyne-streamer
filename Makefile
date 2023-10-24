test:
	go test -cover -coverprofile=coverprofile -v ./...

.ONE_SHELL:
citest:
	jackd -r -d dummy &>/dev/null &
	sleep 1
	$(MAKE) test
