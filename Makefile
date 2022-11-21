# The Go compiler to use
GOC = go

# Set go's environment for building against Linux targets
export GOOS        = linux
export GOARCH      = amd64

# Configure this release's verision and commit
# VERSION = $(shell git describe --tags)
# CCOMMIT = $(shell git rev-parse --short HEAD)

BIN_DIR = ./bin

SOURCES = $(wildcard dvnet/*.go) main.go
TRASH = $(addprefix $(BIN_DIR)/,dvnet)

# The `-ldflags` option lets us define global variables at compile time!
# Check https://stackoverflow.com/questions/11354518/application-auto-build-versioning
# for more information on that!
$(BIN_DIR)/dvnet: $(SOURCES)
	@#echo "Building commit $(CCOMMIT) for version $(VERSION)"
	@$(GOC) build -o $@ -ldflags "-X main.version=$(VERSION) -X main.commit=$(CCOMMIT)"

.PHONY: clean

clean:
	@rm $(TRASH)
