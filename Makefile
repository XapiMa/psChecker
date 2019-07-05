all: webStatusChecker

psChecker:
	GOOS=linux go build -ldflags '-w -s -extldflags "-static"' -o $@ ./cmd/psChecker
