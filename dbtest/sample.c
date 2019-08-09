#include <stdio.h>
#include <string.h>
#include <assert.h>
#include "sqlite3.h"

int main(int argc, char *argv[]) {
  printf("Hello.\n");

  sqlite3 *db;
  sqlite3_stmt *stmt;
  int f;

  f = sqlite3_open("test.db", &db);
  if (f != SQLITE_OK) {
    printf("Error opening db (%s).\n", sqlite3_errmsg(db));
    return 1;
  }

  //char *sql = "SELECT id, alias, desc FROM user WHERE alias = ?1";
  char *sql = "SELECT id, alias, desc FROM user WHERE alias = @alias";

  f = sqlite3_prepare_v2(db, sql, -1, &stmt, 0);
  if (f != SQLITE_OK) {
    printf("Error preparing sql (%s).\n", sqlite3_errmsg(db));
    return 1;
  }

  //sqlite3_bind_text(stmt, 1, "lky", strlen("lky"), SQLITE_STATIC);
  int ialias = sqlite3_bind_parameter_index(stmt, "@alias");
  sqlite3_bind_text(stmt, ialias, "lky", strlen("lky"), SQLITE_STATIC);

  printf("--- Query Results ---\n");
  while (sqlite3_step(stmt) != SQLITE_DONE) {
    const char *id = sqlite3_column_text(stmt, 0);
    const char *alias = sqlite3_column_text(stmt, 1);
    const char *desc = sqlite3_column_text(stmt, 2);

    printf("%s\t%s\t%s\n", id, alias, desc);
  }
  sqlite3_finalize(stmt);

  sqlite3_close(db);
  return 0;
}

sqlite3_stmt *_create_table = NULL;

void init_stmts(sqlite3 *db) {
  int rc;
  char *sql;

  sql = "CREATE TABLE IF NOT EXISTS ?1 (id TEXT PRIMARY KEY, pubdate TEXT, moddate TEXT, title TEXT, body TEXT)";
  rc = sqlite3_prepare_v2(db, sql, -1, &_create_table, 0);
  assert(rc == SQLITE_OK);
}

void init_table(sqlite3 *db, char *tbl) {
}

