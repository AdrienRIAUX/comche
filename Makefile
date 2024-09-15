# Define development test flags
TEST_FLAGS = -mode="root" -dir="./examples"

# Build binary file
build:
	go build -o bin/main main.go

# Run main file
run:
	go run main.go

# Run main file with dev test flags
dev_run:
	go run main.go $(TEST_FLAGS)
