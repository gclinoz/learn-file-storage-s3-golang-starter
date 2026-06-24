.PHONY: all clean

all:
	go build -o out

clean:
	rm ./out
