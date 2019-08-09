// Utility functions
//
// Functions
// ---------
// Malloc(size_t n) (void *)
//

#include <stdlib.h>
#include <stdio.h>
#include <stdarg.h>

void *Malloc(size_t n) {
    return malloc(n);
}

void Free(void *p) {
    free(p);
}

void Log(const char *fmt, ...) {
    char buf[1000];
    va_list args;
    va_start(args, fmt);
    vsnprintf(buf, 1000, fmt, args);
    va_end(args);

    fprintf(stderr, "%s", buf);
    fprintf(stderr, "\n");
}


