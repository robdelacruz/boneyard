#ifndef __UTIL__
#define __UTIL__

#include <stdlib.h>

void *Malloc(size_t n);
void Free(void *p);
void Log(const char *fmt, ...);

#endif

