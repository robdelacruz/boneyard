#ifndef __MAP__
#define __MAP__

#include <stdlib.h>
#include <string.h>

typedef struct MapNode {
  char *k;
  char *v;
  struct MapNode *next;
} MapNode;

typedef struct Map {
  MapNode *head;
} Map;

typedef struct MapIter {
  MapNode *head;
  MapNode *cur;
} MapIter;

Map *map_new();
void map_destroy(Map *map);
void map_set(Map *map, char *k, char *v);
char *map_get(Map *map, char *k);
void map_del(Map *map, char *k);

MapIter *mapiter_new(Map *map);
void mapiter_destroy(MapIter *iter);
int mapiter_next(MapIter *iter);
char *mapiter_k(MapIter *iter);
char *mapiter_v(MapIter *iter);

#endif

