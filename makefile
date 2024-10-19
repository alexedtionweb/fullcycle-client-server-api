# Go bin
GO = go

SERVER_BINARY = build/server
CLIENT_BINARY = build/client

all: server client

server:
	go build -o $(SERVER_BINARY) server/main.go

client:
	go build -o $(CLIENT_BINARY) client/main.go

run-server: server
	@echo "Starting server..."
	./$(SERVER_BINARY)

# Run the client
run-client: client
	@echo "Starting client..."
	./$(CLIENT_BINARY)


# Clean up built binaries
clean:
	rm -f $(SERVER_BINARY) $(CLIENT_BINARY)


.PHONY: all server client run-server run-client run clean