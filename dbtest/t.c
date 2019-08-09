#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <assert.h>
#include <ctype.h>
#include "sqlite3/sqlite3.h"
#include "map.h"
#include "list.h"

void init_stmts(sqlite3 *db);
void destroy_stmts();
void init_table(sqlite3 *db, char *tbl);
List *query_nodes(sqlite3 *db, char *tbl, char *where);
void print_node(Map *node);

int main(int argc, char *argv[]) {
  sqlite3 *db;
  int rc;

  rc = sqlite3_open("test.db", &db);
  assert(rc == SQLITE_OK);

  init_table(db, "nbdata");

  List *nodes = query_nodes(db, "nbdata", NULL);
  printf("List nodes, num nodes = %d:\n", list_len(nodes));
  for (int i=0; i < list_len(nodes); i++) {
    Map *n = list_get(nodes, i);
    print_node(n);
    printf("-------------------\n");
  }
  list_destroy(nodes);

  sqlite3_close(db);

  // Test map
  Map *map = map_new();
  map_set(map, "first", "First node");
  map_set(map, "rob", "Rob de la Cruz");
  map_set(map, "lky", "Lee Kuan Yew");
  map_set(map, "rob", "Overwritten Rob de la Cruz");
  map_set(map, "del123", "Something to delete");
  map_set(map, "lhl", "Lee Hsien Loong");

  Map *emptymap = map_new();

  char *v1 = map_get(map, "rob");
  char *v2 = map_get(map, "lky");
  char *v3 = map_get(map, "lky2");
  char *v4 = map_get(map, "del123");
  char *v5 = map_get(map, "first");
  printf("rob = '%s'\nlky = '%s'\nlky2 = '%s'\n", v1, v2, v3);
  printf("del123 = '%s'\nfirst = '%s'\n", v4, v5);

  map_del(map, "del123");
  map_del(map, "nonexisting");
  map_del(map, "first");
  map_del(map, "del123");

  printf("Iterating through map...\n");
  MapIter *iter = mapiter_new(map);
  while (mapiter_next(iter)) {
    printf("k = '%s', v = '%s'\n", mapiter_k(iter), mapiter_v(iter));
  }
  mapiter_destroy(iter);

  map_destroy(map);
  map_destroy(emptymap);

  // Test list
  List *list = list_new(NULL);
  list_append(list, strdup("abc"));
  list_append(list, strdup("def"));
  list_append(list, strdup("ghi"));

  printf("List len = %d, cap = %d\n", list_len(list), list->cap);
  for (int i=0; i < list_len(list); i++) {
    char *item = list_get(list, i);
    printf("[%d] = '%s'\n", i, item);
  }

  list_destroy(list);

  return 0;
}

// Count the total number of chars in all the format specification tags
// (Ex. "%10s", "%d", "%5.5d") in a printf-style format string. 
int count_fspec(char *fstr) {
  assert(fstr);

  // Sample fstr inputs:
  // "blah blah %s %10s %2d%5.34f blah 110%% %d%%%s"

  int ntotal = 0;

  int fInTag = 0;
  int lenfstr = strlen(fstr);
  for (int i=0; i < lenfstr; i++) {
    // Get current char and next char (lookahead).
    char ch = fstr[i];
    char nextch = 0;
    if (i < lenfstr-1) {
      nextch = fstr[i+1];
    }

    // Skip over any "%%", which also terminates a tag.
    if (ch == '%' && nextch == '%') {
      fInTag = 0;
      i++;
      continue;
    }
    // Any whitespace termininates a %nnn tag.
    if (isspace(ch)) {
      fInTag = 0;
      continue;
    }
    // Any non-whitespace char following a '%' adds to tag count.
    if (fInTag) {
      ntotal++;
      continue;
    }
    // '%' char starts a tag
    if (ch == '%') {
      fInTag = 1;
      ntotal++;
      continue;
    }
    // Ignore any char not in tag.
  }

  return ntotal;
}

char *sqlstr_new(char *sqlf, va_list args) {
  va_list tmpArgs;
  va_copy(tmpArgs, args);

  int lenArgs = 0;
  char *arg;
  while ((arg = va_arg(tmpArgs, char*)) != NULL) {
    lenArgs += strlen(arg);
    printf("lenArgs=%d, arg='%s'\n", lenArgs, arg);
  }
  va_end(tmpArgs);

  int sqlLen = strlen(sqlf) - count_fspec(sqlf) + lenArgs + 1;
  char *sql = malloc(sqlLen);
  int nUsed = vsnprintf(sql, sqlLen, sqlf, args);
  assert(nUsed < sqlLen);

  printf("sql = '%s'\n", sql);
  return sql;
}

sqlite3_stmt *stmt_new(sqlite3 *db, char *sqlf, ...) {
  va_list args;
  va_start(args, sqlf);
  char *sql = sqlstr_new(sqlf, args);
  va_end(args);

  sqlite3_stmt *stmt;
  int rc = sqlite3_prepare_v2(db, sql, -1, &stmt, NULL);
  assert(rc == SQLITE_OK);

  free(sql);
  return stmt;
}

// Create node table and related objects.
void init_table(sqlite3 *db, char *tbl) {
  char *sqlf = "CREATE TABLE IF NOT EXISTS %s (id TEXT PRIMARY KEY, pubdate TEXT, moddate TEXT, title TEXT, body TEXT)";

  sqlite3_stmt *stmt = stmt_new(db, sqlf, tbl, NULL);
  sqlite3_step(stmt);
  sqlite3_finalize(stmt);
}

List *query_nodes(sqlite3 *db, char *tbl, char *where) {
  assert(tbl);

  if (where == NULL) {
    where = "1=1";
  }

  char *sqlf = "SELECT id, pubdate, moddate, title, body FROM %s WHERE %s";
  sqlite3_stmt *stmt = stmt_new(db, sqlf, tbl, where, NULL);

  List *nodes = list_new((DestroyCB)map_destroy);
  while (sqlite3_step(stmt) != SQLITE_DONE) {
    Map *node = map_new();
    char *id      = (char*) sqlite3_column_text(stmt, 0);
    char *pubdate = (char*) sqlite3_column_text(stmt, 1);
    char *moddate = (char*) sqlite3_column_text(stmt, 2);
    char *title   = (char*) sqlite3_column_text(stmt, 3);
    char *body    = (char*) sqlite3_column_text(stmt, 4);

    map_set(node, "id", id);
    map_set(node, "pubdate", pubdate ? pubdate : "");
    map_set(node, "moddate", moddate ? moddate : "");
    map_set(node, "title", title ? title : "");
    map_set(node, "body", body ? body : "");

    list_append(nodes, node);
  }

  sqlite3_finalize(stmt);
  return nodes;
}

void print_node(Map *node) {
  MapIter *iter = mapiter_new(node);

  while (mapiter_next(iter)) {
    printf("%s\t:%s\n", mapiter_k(iter), mapiter_v(iter));
  }

  mapiter_destroy(iter);
}


