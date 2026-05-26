# ------- APP

init:
	@echo - Installing dependencies ...
	go mod download

app:
	@go build -o ./build/app ./cmd/app
	@./build/app

test\:help:
	@printf "\
	\n\
	To run all tests:\
	\n\
		make test\
	\n\n\
	To test a path:\
	\n\
		make test path='path/to/your_test(s)'\
	\n\n\
	To test a scope in tests directory:\
	\n\
		make test scope=server\
	\n\n\
	To test in race mode:\
	\n\
		make test race=t\
	\n\n\
	To run tests by a filter:\
	\n\
		make test filter='Wallet'\
	\n\n\
	To run tests x times:\
	\n\
		make test count=10\
	\n\n\
	To run tests with detail:\
	\n\
		make test verbose=t\
	"

test:
	@go test -p 1 \
		$(if $(path), $(path), $(if $(scope), ./tests/$(scope)/*, ./...)) \
		$(if $(race), -race) \
		$(if $(verbose), -v) \
		-count $(if $(count), $(count), 1) \
		-run ^.*$(filter).* \
		| grep -a --line-buffered -v -e "no test files" -e "no tests to run"
