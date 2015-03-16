TARGET=drumcli

$(TARGET): cli/cli.go
	go build -v -o $@ ./cli

.PHONY: clean
clean:
	rm -v $(TARGET)
