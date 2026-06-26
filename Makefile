.PHONY: all clean

all:
	go build -o out

clean:
	rm ./out
	@curl -X POST http://localhost:8091/admin/reset
