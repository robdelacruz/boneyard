import datetime
import re
import sqlite3
import os
import pathlib

import util

def connect_db():
    """ Return db connection """
    homedir = str(pathlib.Path.home())
    dbfilename = "nbdata.db"
    dbpath = os.path.join(homedir, dbfilename)

    con = sqlite3.connect(dbpath)
    con.row_factory = sqlite3.Row
    return con


def init_table(con, tbl):
    """ Create data table if it doesn't exist yet. """

    sql = """CREATE TABLE IF NOT EXISTS %s (
_id TEXT PRIMARY KEY,
_body TEXT)""" % tbl

    con.cursor().execute(sql)
    con.commit()

    init_fts_table(con, tbl)


def init_fts_table(con, tbl):
    """ Create full-text search table and triggers if they don't exist yet. """

    sqls = [
        f"CREATE VIRTUAL TABLE IF NOT EXISTS {tbl}_fts USING fts5(__id, __body, tokenize = porter)",
        f"CREATE TRIGGER IF NOT EXISTS {tbl}_insert AFTER INSERT ON {tbl} BEGIN INSERT INTO {tbl}_fts (__id, __body) VALUES (new._id, new._body); END",
        f"CREATE TRIGGER IF NOT EXISTS {tbl}_update AFTER UPDATE OF _body ON {tbl} BEGIN UPDATE {tbl}_fts SET __body = new._body WHERE __id = old._id; END",
        f"CREATE TRIGGER IF NOT EXISTS {tbl}_delete AFTER DELETE ON {tbl} BEGIN DELETE FROM {tbl}_fts WHERE __id = old._id; END"
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)
    con.commit()


def purge_table(con, tbl):
    sqls = [
        f"DROP TABLE IF EXISTS {tbl}",
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)
    con.commit()

    purge_fts_table(con, tbl)


def purge_fts_table(con, tbl):
    sqls = [
        f"DROP TABLE IF EXISTS {tbl}_fts",
        f"DROP TRIGGER IF EXISTS {tbl}_insert",
        f"DROP TRIGGER IF EXISTS {tbl}_update",
        f"DROP TRIGGER IF EXISTS {tbl}_delete"
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)
    con.commit()


def rebuild_fts_table(con, tbl):
    """ Clear and repopulate full-text search table from data table. """

    purge_fts_table(con, tbl)
    init_fts_table(con, tbl)

    sql = f"INSERT INTO {tbl}_fts (__id, __body) SELECT _id, _body FROM {tbl}"
    con.cursor().execute(sql)
    con.commit()


def table_cols(con, tbl):
    """ Return set containing the table's columns. """

    cur = con.cursor()
    sql = """SELECT sql FROM sqlite_master WHERE tbl_name = ?"""
    cur.execute(sql, (tbl,))
    row = cur.fetchone()

    if row == None:
        return []
    createtbl_sql = row[0]

    # Get field definition
    # the contents of '...' in CREATE TABLE tblname (...)
    m = re.search(r'(?s)\(\s*(.*)\s*\)', createtbl_sql)
    if m == None:
        return []
    fields_sql = m.group(1).strip()

    # Separate each field definition.
    # _id TEXT PRIMARY KEY,
    # _body TEXT
    # field1 TEXT
    # field2 TEXT
    field_defns = re.split(r',\s*', fields_sql)

    # First word of each field definition is the field name.
    # _id
    # _body
    # field1
    # field2
    cols = set()
    for field_defn in field_defns:
        field = re.split(r'\s+', field_defn)
        cols.add(field[0])
    return cols


def all_tables(con):
    """ Return set of all tables in db """

    cur = con.cursor()
    sql = "SELECT DISTINCT tbl_name FROM sqlite_master WHERE tbl_name NOT LIKE '%\_fts%' ESCAPE '\\' ORDER BY tbl_name"
    cur.execute(sql)

    tbls = set()
    for row in cur:
        tbls.add(row[0])
    return tbls


def add_table_cols(con, tbl, cols):
    """ Add new columns to table
    con: sqlite3 connection
    tbl: db table
    cols: set of column names to add

    Only new cols not existing in the table will be added.
    """

    existing_cols = table_cols(con, tbl)
    new_cols = cols.difference(existing_cols)
    for new_col in new_cols:
        sql = """ALTER TABLE %s ADD %s""" % (tbl, new_col)
        con.cursor().execute(sql)
    con.commit()


def replace_vars(s):
    """ Replace @@varx with instance values """
    today_iso = datetime.datetime.now().isoformat()
    s = s.replace('@@today', today_iso)

    return s


def rec_new():
    return {'_id': "", '_body': ""}


def is_valid_field_name(field):
    # Valid field name starts with alphanumeric char followed by 0 or more
    # alphanumeric or '_' chars.
    if re.search(r'^[A-Za-z0-9]\w*$', field):
        return True
    return False


def rec_from_mktext(mktext, id=None):
    """
    Ex.
    mktext:
        line 1
        line 2
        .amt 123.45
        line 3
        .author rob

    returns: rec where
        rec['_id']: <id>
        rec['_body']:
        line 1
        line 2
        line 3
        rec['amt']: 123.45
        rec['author']: rob
    """

    body_lines = mktext.split('\n')
    if len(body_lines) == 0:
        return

    rec = {}
    new_body_lines = []
    for line in body_lines:
        # Match .field1 val1
        # Valid field name starts with alphanumeric char followed by 0 or more
        # alphanumeric or '_' chars.
        m = re.search(r'^\.([A-Za-z0-9]\w*)\s+(.*)$', line)
        if m == None:
            new_body_lines.append(line)
            continue

        k = m.group(1)
        v = m.group(2)
        v = replace_vars(v)
        rec[k] = v

    rec['_body'] = "\n".join(new_body_lines).strip() + "\n"
    rec['_id'] = id or ""

    return rec


def mktext_from_rec(rec):
    """ Return textual representation of rec dotfields.
    """

    # Get list of dot fields in alphabetical order.
    #   dot fields = [all rec fields] - [base fields]
    base_fields = {'_id', '_body'}
    dot_fields = sorted(set(rec.keys()).difference(base_fields))

    dotfield_lines = []
    for field in dot_fields:
        line = ".%s %s" % (field, rec[field])
        dotfield_lines.append(line)

    body = rec.get('_body', "")
    field_part = '\n'.join(dotfield_lines)

    if not body:
        return field_part

    return (body + "\n" + field_part)


def update_rec(con, tbl, rec):
    """ Add a new rec to table.
    con: sqlite connection
    tbl: table name
    rec: dict containing rec fields
    """

    # Create table if it doesn't exist.
    if not tbl in all_tables(con):
        init_table(con, tbl)

    # If existing rec, delete rec first before insert.
    id = rec.get('_id')
    if id != None:
        sql = """DELETE FROM %s WHERE _id = ?""" % (tbl)
        con.cursor().execute(sql, [id])

    # Make sure table schema has all rec columns needed.
    # Create new columns if necessary.
    add_table_cols(con, tbl, set(rec.keys()))

    # Generate new id if not specified
    if id == None or id == "":
        rec['_id'] = util.gen_id()

    rec_cols = list(rec.keys())      # ['_id', '_body', 'field1', 'field2']
    qs = ['?'] * len(rec_cols)       # ['?', '?', '?', '?']

    sql_cols = ", ".join(rec_cols)   # Ex. "_id, _body, field1, field2"
    q_cols   = ", ".join(qs)         # Ex. "?, ?, ?, ?"
    sql = """INSERT INTO %s (%s) VALUES (%s)""" % (tbl, sql_cols, q_cols)

    vals = []
    for k in rec_cols:
        vals.append(rec[k])

    con.cursor().execute(sql, vals)
    con.commit()

    return rec['_id']


def table_cols_sorted(con, tbl):
    """ Return list containing the table's columns with base fields first followed by the
        rest of the fields in alphabetical order.
        Ex. ['_id', '_body', 'amt', 'author', 'footnote]"""
    cols = list(table_cols(con, tbl))
    cols.remove('_id')
    cols.remove('_body')
    cols.sort()
    cols = ['_id', '_body'] + cols
    return cols


def read_rec(con, tbl, id):
    """ Read one rec of specified id.
    con: sqlite connection
    tbl: table name
    id: rec id
    """

    if not tbl in all_tables(con):
        return None

    # select cols always starts with '_id', '_body' followed by the rest of the fields
    # in alphabetical order
    # Ex. "_id, _body, amt, author, footnote"
    scols = ", ".join(table_cols_sorted(con, tbl))

    sql = f"SELECT {scols} FROM {tbl} WHERE _id = ?"
    cur = con.cursor()
    cur.execute(sql, [id])

    rec = None
    for row in cur:
        rec = {}
        for k in row.keys():
            if row[k] != None:
                rec[k] = row[k]
        break

    return rec


def read_recs(con, tbl, q="", qwhere="", qorder=""):
    """ Read recs matching where clause.
    con: sqlite connection
    tbl: table name
    qwhere: sql where clause
    """

    if qwhere == "":
        qwhere = "1"
    if qorder == "":
        qorder = "_id"

    if not tbl in all_tables(con):
        return []

    # select cols always starts with '_id', '_body' followed by the rest of the fields
    # in alphabetical order
    # Ex. "_id, _body, amt, author, footnote"
    scols = ", ".join(table_cols_sorted(con, tbl))

    if q != "":
        sql = f"SELECT {scols} FROM {tbl} AS d INNER JOIN {tbl}_fts AS fts ON d._id = fts.__id WHERE {tbl}_fts MATCH '{q}' AND {qwhere} ORDER BY {qorder}"
    else:
        sql = f"SELECT {scols} FROM {tbl} WHERE {qwhere} ORDER BY {qorder}"
    cur = con.cursor()

    try:
        cur.execute(sql)
    except:
        return []

    recs = []
    for row in cur:
        rec = {}
        for k in row.keys():
            if row[k] != None:
                rec[k] = row[k]
        recs.append(rec)

    return recs


def del_rec(con, tbl, id):
    """ Delete one rec of specified id.
    con: sqlite connection
    tbl: table name
    id: rec id
    """

    if not tbl in all_tables(con):
        return None

    sql = """DELETE FROM %s WHERE _id = ?""" % tbl
    con.cursor().execute(sql, [id])
    con.commit()


def main():
    con = connect_db()
    init_table(con, "nbdata")
    init_table(con, "robtable")
    init_table(con, "tblx")

    tbls = all_tables(con)
    add_table_cols(con, "tblx", {'fld1', 'fld2', 'amt'})
    cols = table_cols(con, "nbdata")
    cols = table_cols(con, "tblx")

    #purge_table(con, "nbdata")
    #init_table(con, "nbdata")
    #rebuild_fts_table(con, "nbdata")

    mktext = """Fortune favors the bold.
.author rob3
.topic testing
.dt @@today
"""

    print("Updating rec...")
    rec = rec_from_mktext(mktext)
    rec['_id'] = "123"
    new_id = update_rec(con, "nbdata", rec)
    print(f"New rec: {new_id}")

    #print(f"Reading rec {new_id}...")
    #rec = read_rec(con, "nbdata", new_id)
    #print(repr(rec))

    q = "please"
    #qwhere = "dt >= '2019-02-27'"
    qwhere = ""
    print(f"Reading recs where {qwhere}")
    recs = read_recs(con, "nbdata", q, qwhere)
    for rec in recs:
        print(repr(rec))

    #print("Deleting rec 123...")
    #del_rec(con, "nbdata", "123")


if __name__ == "__main__":
    main()

