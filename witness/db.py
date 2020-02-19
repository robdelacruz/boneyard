import re
import sqlite3
import os
import pathlib
from enum import IntEnum

import conv

class EntryType(IntEnum):
    LOG = 0
    NOTE = 1
    EVENT = 2
    TASK = 3

class StatusType(IntEnum):
    OPEN = 0
    COMPLETED = 1

def connect_db(dbfile=None):
    """ Return db connection """

    # Get source database filename path (dbfile) and directory (dbdir).
    dbdir = ""
    if not dbfile:
        # Use ~/.witness/data.db by default if no file specified.
        dbdir = os.path.join(str(pathlib.Path.home()), ".witness")
        dbfile = os.path.join(dbdir, "data.db")
    else:
        dbdir = os.path.dirname(dbfile)

    # Create dbdir if needed.
    # dbfile will be created by sqlite3.connect() if it doesn't yet exist.
    if dbdir != "":
        os.makedirs(dbdir, exist_ok=True)

    print(f"dbfile = '{dbfile}'")
    con = sqlite3.connect(dbfile)
    con.row_factory = sqlite3.Row
    return con


def commit(con):
    con.commit()


def exec_sqls(con, sqls):
    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)


def purge_tables(con):
    sqls = [
        "PRAGMA foreign_keys = ON;",
        "DROP TABLE IF EXISTS entry;",
        "DROP TABLE IF EXISTS topic;",
]
    exec_sqls(con, sqls)


def init_tables(con):
    sqls = [
    """CREATE TABLE IF NOT EXISTS entry (
        entry_id INTEGER NOT NULL PRIMARY KEY,
        entry_type INTEGER DEFAULT 0,
        entry_createdt TEXT DEFAULT (date('now')),
        entry_body TEXT DEFAULT '',
        entry_startdt TEXT NULL,
        entry_enddt TEXT NULL,
        entry_status INTEGER DEFAULT 0,
        topic_id INTEGER DEFAULT 0,
        FOREIGN KEY(topic_id) REFERENCES topic
    );""",
    """CREATE TABLE IF NOT EXISTS topic (
        topic_id INTEGER NOT NULL PRIMARY KEY,
        topic_name TEXT NOT NULL
    );""",
    "INSERT OR REPLACE INTO topic (topic_id, topic_name) VALUES (0, 'Unassigned');"
]
    exec_sqls(con, sqls)


def add_entry(con, entry):
    cur = con.cursor()

    entry_id = entry.get("id")
    entry_type = entry.get("type", EntryType.LOG)
    createdt = entry.get("createdt", conv.today_isodt())
    body = entry.get("body", "")
    startdt = entry.get("startdt")
    enddt = entry.get("enddt")
    status = entry.get("status", 0)
    topic_id = entry.get("topic_id", 0)

    if entry_type == EntryType.EVENT or entry_type == EntryType.TASK:
        if not startdt and not enddt:
            startdt = enddt = createdt
        elif not startdt:
            startdt = enddt
        elif not enddt:
            enddt = startdt

    sql = "INSERT OR REPLACE INTO entry (entry_id, entry_type, entry_createdt, entry_body, entry_startdt, entry_enddt, entry_status, topic_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
    cur.execute(sql, (entry_id, entry_type, createdt, body, startdt, enddt, status, topic_id))

ENTRY_COLS = "entry_id as id, entry_type as type, entry_createdt as createdt, entry_body as body, entry_startdt as startdt, entry_enddt as enddt, entry_status as status, topic_id"

def read_entry(con, entry_id):
    sql = f"""SELECT {ENTRY_COLS}
    FROM entry
    WHERE entry_id = ?
"""
    cur = con.cursor()
    cur.execute(sql, [entry_id])

    for row in cur:
        entry = {}
        for k in row.keys():
            if row[k] != None:
                entry[k] = row[k]
        return entry
    return None


def read_entries(con, where, orderby, parms):
    sql = f"SELECT {ENTRY_COLS} FROM entry WHERE {where} ORDER BY {orderby}"
    cur = con.cursor()
    cur.execute(sql, parms)

    entries = []
    for row in cur:
        entry = {}
        for k in row.keys():
            if row[k] != None:
                entry[k] = row[k]
        entries.append(entry)

    return entries


def read_today_all_entries(con, dt=conv.today_isodt()):
    where = """
        (entry_type IN (0, 1) AND entry_createdt = :dt) OR
        (entry_type = 2 AND :dt BETWEEN entry_startdt AND entry_enddt) OR
        (entry_type = 3 AND :dt BETWEEN entry_startdt AND entry_enddt)
    """
    orderby = "entry_type desc, entry_enddt, entry_startdt, entry_id"
    parms = {"dt": dt}

    return read_entries(con, where, orderby, parms)


def read_today_logs(con, dt=conv.today_isodt()):
    where = "entry_type IN (0) AND entry_createdt = :dt"
    orderby = "entry_type, entry_id"
    parms = {"dt": dt}

    return read_entries(con, where, orderby, parms)


def read_today_events(con, dt=conv.today_isodt()):
    where = "entry_type = 2 AND :dt BETWEEN entry_startdt AND entry_enddt"
    orderby = "entry_enddt, entry_startdt, entry_id"
    parms = {"dt": dt}
    return read_entries(con, where, orderby, parms)


def read_today_tasks(con, dt=conv.today_isodt()):
    where = "entry_type = 3 AND :dt BETWEEN entry_startdt AND entry_enddt"
    orderby = "entry_enddt, entry_startdt, entry_id"
    parms = {"dt": dt}
    return read_entries(con, where, orderby, parms)


def read_recent_notes(con, dt=conv.today_isodt()):
    where = "entry_type = 1"
    orderby = "entry_createdt desc limit 10"
    parms = {}
    return read_entries(con, where, orderby, parms)


def del_entry(con, entry_id):
    sql = "DELETE FROM entry WHERE entry_id = ?"
    cur = con.cursor()
    cur.execute(sql, [entry_id])


def main():
    con = connect_db("test.db")
    purge_tables(con)
    init_tables(con)
    commit(con)

    createdt = "2019-04-01"

    add_entry(con, {"type": EntryType.LOG, "createdt": createdt, "body": "Log 1"})
    add_entry(con, {"type": EntryType.LOG, "createdt": createdt, "body": "Log 2"})
    add_entry(con, {"type": EntryType.NOTE, "createdt": createdt, "body": "Note 1"})
    add_entry(con, {"type": EntryType.NOTE, "createdt": createdt, "body": "Note 2"})
    add_entry(con, {"type": EntryType.NOTE, "createdt": createdt, "body": "Note 3"})
    add_entry(con, {"type": EntryType.EVENT, "createdt": createdt, "body": "April Fool's Day", "startdt": "2019-04-01"})
    add_entry(con, {"type": EntryType.EVENT, "createdt": createdt, "body": "April iteration 1", "startdt": "2019-04-01", "enddt": "2019-04-05"})
    add_entry(con, {"type": EntryType.EVENT, "createdt": createdt, "body": "April iteration 2", "startdt": "2019-04-06", "enddt": "2019-04-15"})
    add_entry(con, {"type": EntryType.TASK, "createdt": createdt, "body": "Pay April bills", "startdt": "2019-04-09", "enddt": "2019-04-10"})
    add_entry(con, {"type": EntryType.TASK, "createdt": createdt, "body": "Install updates", "startdt": "2019-04-10"})
    commit(con)

    entries = read_today_all_entries(con, createdt)
    print(f"read_today_all_entries() from '{createdt}':")
    for entry in entries:
        print(entry)


if __name__ == "__main__":
    main()

