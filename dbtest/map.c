#include <assert.h>
#include "map.h"

#define MAP_MAXLEN_K 500
#define MAP_MAXLEN_V 5000

// Create a new map.
Map *map_new() {
  Map *map = malloc(sizeof(Map));
  map->head = NULL;
}

// Deallocate all map nodes and key/values.
void map_destroy(Map *map) {
  assert(map);

  MapNode *n = map->head;
  while (n != NULL) {
    MapNode *tmpNext = n->next;

    free(n->k);
    free(n->v);
    free(n);

    n = tmpNext;
  }
}

MapNode *mapnode_new(char *k, char *v) {
  MapNode *node = malloc(sizeof(MapNode));
  node->k = strdup(k);
  node->v = strdup(v);
  node->next = NULL;

  return node;
}

// Set key and value to map.
void map_set(Map *map, char *k, char *v) {
  assert(map);
  assert(k);
  assert(v);
  assert(strlen(k) <= MAP_MAXLEN_K);  // be suspicious if k,v too long,
  assert(strlen(v) <= MAP_MAXLEN_V);  // might be corrupted memory.

  // Empty map, add as first node.
  if (map->head == NULL) {
    map->head = mapnode_new(k, v);
    return;
  }

  // Iterate through nodes.
  MapNode *n = map->head;
  while (n->next != NULL) {
    assert(n->k);
    assert(n->v);

    // Key already exists, so overwrite value.
    if (strcmp(n->k, k) == 0) {
      free(n->v);
      n->v = strdup(v);
      return;
    }

    n = n->next;
  }

  // Nonexistent key so add new node at the end.
  n->next = mapnode_new(k, v);
}

// Lookup k in map. Return empty string "" if not found.
char *map_get(Map *map, char *k) {
  assert(map);
  assert(k);

  MapNode *n = map->head;
  while (n != NULL) {
    assert(n->k);
    assert(n->v);

    // Found
    if (strcmp(n->k, k) == 0) {
      return n->v;
    }

    n = n->next;
  }

  // Not found
  return "";
}

void map_del(Map *map, char *k) {
  assert(map);
  assert(k);

  MapNode *nprev = NULL;
  MapNode *n = map->head;
  while (n != NULL) {
    assert(n->k);
    assert(n->v);

    // Found? Remove node n
    //   nprev => n => n->next
    if (strcmp(n->k, k) == 0) {
      if (nprev == NULL) {
        map->head = n->next;
      } else {
        nprev->next = n->next;
      }

      free(n->k);
      free(n->v);
      free(n);
      break;
    }

    nprev = n;
    n = n->next;
  }
}

MapIter *mapiter_new(Map *map) {
  assert(map);

  MapIter *iter = malloc(sizeof(MapIter));
  iter->head = map->head;
  iter->cur = NULL;
}

void mapiter_destroy(MapIter *iter) {
  free(iter);
}

int mapiter_next(MapIter *iter) {
  assert(iter);

  // Empty map?
  if (iter->head == NULL) {
    return 0;
  }

  // Starting out? Advance to first node.
  if (iter->cur == NULL) {
    iter->cur = iter->head;
    return 1;
  }

  // No more nodes?
  if (iter->cur->next == NULL) {
    return 0;
  }

  // Go to next node.
  iter->cur = iter->cur->next;
  return 1;
}

char *mapiter_k(MapIter *iter) {
  assert(iter);
  assert(iter->cur);

  if (iter->cur == NULL) {
    return "";
  }
  return iter->cur->k;
}

char *mapiter_v(MapIter *iter) {
  assert(iter);
  assert(iter->cur);

  if (iter->cur == NULL) {
    return "";
  }
  return iter->cur->v;
}

