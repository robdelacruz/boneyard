nb: t.c map.c list.c sqlite3.o
	gcc -o nb t.c map.c list.c sqlite3.o

sqlite3.o: sqlite3/sqlite3.c
	gcc -c -DSQLITE_THREADSAFE=0 -DSQLITE_OMIT_LOAD_EXTENSION sqlite3/sqlite3.c

clean:
	rm nb sqlite3.o

