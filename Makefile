.PHONY: all clean

all:
	go build -o out

clean:
	@curl -X POST http://localhost:8091/admin/reset
	rm ./out
