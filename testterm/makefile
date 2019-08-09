all: t

t: t.c term.c gapbuf.c
	$(CC) -o t t.c term.c gapbuf.c -Wall -W -pedantic -std=c99

str: str.c
	$(CC) -D TEST -o str str.c

gapbuf: gapbuf.c
	$(CC) -D TEST -o gapbuf gapbuf.c

clean:
	rm t str gapbuf

