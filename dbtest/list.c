#include <stdlib.h>
#include <assert.h>
#include "list.h"

List *list_new(void (*free_item)(void*)) {
  List *l = malloc(sizeof(List));
  l->len = 0;
  l->free_item = free_item;

  // Initial capacity size is 1 element.
  l->cap = 1;
  l->items = malloc(l->cap * sizeof(void*));
}

void list_destroy(List *l) {
  assert(l);

  // Call user-supplied free function on each item.
  if (l->free_item != NULL) {
    for (int i=0; i < list_len(l); i++) {
      void *item = list_get(l, i);
      l->free_item(item);
    }
  }

  free(l->items);
  free(l);
}

void list_append(List *l, void *item) {
  assert(l);
  assert(item);
  assert(l->len <= l->cap);

  // Increase capacity if needed.
  if (l->len == l->cap) {
    l->cap *= 2;
    l->items = realloc(l->items, l->cap * sizeof(void*));
  }
  assert(l->len < l->cap);

  // Add item to the end.
  l->items[l->len] = item;
  l->len++;
}

unsigned int list_len(List *l) {
  assert(l);
  return l->len;
}

void *list_get(List *l, unsigned int i) {
  assert(l);
  assert(i <= l->len-1);

  return l->items[i];
}

