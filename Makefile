.PHONY: all
all:
	make psChecker_linux
	make psChecker.exe

.PHONY: psChecker_linux
psChecker_linux:
	GOOS=linux go build -ldflags '-w -s -extldflags "-static"' -o $@ ./cmd/psChecker

.PHONY: psChecker.exe
psChecker.exe:
	GOOS=windows go build -ldflags '-w -s -extldflags "-static"' -o $@ ./cmd/psChecker
