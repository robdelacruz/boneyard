#ifndef __LIST__
#define __LIST__

typedef void (*DestroyCB)(void*);

typedef struct List {
  void **items;
  unsigned int len;
  unsigned int cap;
  void (*free_item)(void*);
} List;

List *list_new(void (*free_item)(void*));
void list_destroy(List *l);
void list_append(List *l, void *item);
unsigned int list_len(List *l);
void *list_get(List *l, unsigned int i);

#endif
