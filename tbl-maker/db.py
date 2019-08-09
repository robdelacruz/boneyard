import datetime
import re
import sqlite3
import os
import pathlib

import util

def connect_db(dbfile=None):
    """ Return db connection """

    # Get source database filename path (dbfile) and directory (dbdir).
    dbdir = ""
    if not dbfile:
        # Use ~/.tblmaker/data.db by default if no file specified.
        dbdir = os.path.join(str(pathlib.Path.home()), ".tblmaker")
        dbfile = os.path.join(dbdir, "data.db")
    else:
        dbdir = os.path.dirname(dbfile)

    # Create dbdir if needed.
    # dbfile will be created by sqlite3.connect() if it doesn't yet exist.
    if dbdir != "":
        print(f"dbdir = '{dbdir}'")
        os.makedirs(dbdir, exist_ok=True)

    con = sqlite3.connect(dbfile)
    con.row_factory = sqlite3.Row
    return con


def commit(con):
    con.commit()


def _valid_table_name(tbl):
    invalid_chars = "()[]"
    is_invalid = False
    for c in tbl:
        if c in invalid_chars:
            is_invalid = True
            break

    if is_invalid:
        for c in invalid_chars:
            tbl = tbl.replace(c, "")

    return tbl


def init_table(con, tbl):
    """ Create data table if it doesn't exist yet. """
    tbl = _valid_table_name(tbl)

    sql = f"CREATE TABLE IF NOT EXISTS [{tbl}] (_id TEXT PRIMARY KEY, _body TEXT)"
    con.cursor().execute(sql)

    _init_fts_table(con, tbl)


def _init_fts_table(con, tbl):
    """ Create full-text search table and triggers if they don't exist yet. """
    tbl = _valid_table_name(tbl)

    sqls = [
        f"CREATE VIRTUAL TABLE IF NOT EXISTS [{tbl}_fts] USING fts5(__id, __body, tokenize = porter)",
        f"CREATE TRIGGER IF NOT EXISTS [{tbl}_insert] AFTER INSERT ON [{tbl}] BEGIN INSERT INTO [{tbl}_fts] (__id, __body) VALUES (new._id, new._body); END",
        f"CREATE TRIGGER IF NOT EXISTS [{tbl}_update] AFTER UPDATE OF _body ON [{tbl}] BEGIN UPDATE [{tbl}_fts] SET __body = new._body WHERE __id = old._id; END",
        f"CREATE TRIGGER IF NOT EXISTS [{tbl}_delete] AFTER DELETE ON [{tbl}] BEGIN DELETE FROM [{tbl}_fts] WHERE __id = old._id; END"
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)


def purge_table(con, tbl):
    tbl = _valid_table_name(tbl)
    sqls = [
        f"DROP TABLE IF EXISTS [{tbl}]",
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)

    _purge_fts_table(con, tbl)


def _purge_fts_table(con, tbl):
    tbl = _valid_table_name(tbl)
    sqls = [
        f"DROP TABLE IF EXISTS [{tbl}_fts]",
        f"DROP TRIGGER IF EXISTS [{tbl}_insert]",
        f"DROP TRIGGER IF EXISTS [{tbl}_update]",
        f"DROP TRIGGER IF EXISTS [{tbl}_delete]"
    ]

    cur = con.cursor()
    for sql in sqls:
        cur.execute(sql)


def rebuild_fts_table(con, tbl):
    """ Clear and repopulate full-text search table from data table. """
    tbl = _valid_table_name(tbl)

    _purge_fts_table(con, tbl)
    _init_fts_table(con, tbl)

    sql = f"INSERT INTO [{tbl}_fts] (__id, __body) SELECT _id, _body FROM [{tbl}]"
    con.cursor().execute(sql)


def table_cols(con, tbl):
    """ Return set containing the table's columns. """
    tbl = _valid_table_name(tbl)

    cur = con.cursor()
    sql = f"SELECT sql FROM sqlite_master WHERE tbl_name = ?"
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


def _add_table_cols(con, tbl, cols):
    """ Add new columns to table
    con: sqlite3 connection
    tbl: db table
    cols: set of column names to add

    Only new cols not existing in the table will be added.
    """
    existing_cols = table_cols(con, tbl)
    new_cols = cols.difference(existing_cols)

    # No new table cols to add.
    if len(new_cols) == 0:
        return

    for new_col in new_cols:
        sql = f"ALTER TABLE [{tbl}] ADD {new_col}"
        con.cursor().execute(sql)


def _replace_vars(s):
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


def _convert_to_db_type(sval):
    """ Detect which database type (string, int, float) is most appropriate for sval
        and convert it to its corresponding database type. """

    # floats: Ex. 1.23, 0.23, .23
    if re.search(r'^\s*\d*\.\d+\s*$', sval):
        return float(sval)

    # ints: Ex. 0, 123
    if re.search(r'^\s*\d+\s*$', sval):
        return int(sval)

    # strings: anything else
    return sval


def rec_from_mktext(mktext, id=None):
    """
    Ex.
    mktext:
        line 1
        line 2
        .amt 123.45
        line 3
        .author rob
        .empty_field
        ._field_with_underscore 123
    
    returns: rec where
        rec['_id']: <id>
        rec['_body']:
        line 1
        line 2
        line 3
        rec['amt']: 123.45
        rec['author']: rob

        ".empty_field" line is ignored because there's no value
        "._field_with_underscore" line is ignored because it's an invalid field
            (starts with underscore)
    """

    body_lines = mktext.split('\n')
    if len(body_lines) == 0:
        return

    rec = {}
    new_body_lines = []
    for line in body_lines:
        # Skip over any field definition without a value
        if re.search(r'^\.\w+$', line):
            continue

        # Match field definition  Ex. ".field_1 Value of field_1"
        #   field = "field_1"
        #   val   = "Value of field_1"
        m = re.search(r'^\.(\w+)\s+(.*)$', line)
        if m == None:
            new_body_lines.append(line)
            continue

        k = m.group(1)
        v = m.group(2).strip()

        if not is_valid_field_name(k):
            continue

        if v == "":
            continue

        v = _replace_vars(v)
        v = _convert_to_db_type(v)
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
        line = f".{field} {rec[field]}"
        dotfield_lines.append(line)

    body = rec.get('_body', "")
    field_part = '\n'.join(dotfield_lines)

    if not body:
        return field_part
    if not field_part:
        return body
    return (field_part + "\n\n" + body)


def update_rec(con, tbl, rec):
    """ Add a new rec to table.
    con: sqlite connection
    tbl: table name
    rec: dict containing rec fields
    """
    tbl = _valid_table_name(tbl)

    def _do_update_rec():
        # If existing rec, delete rec first before insert.
        id = rec.get('_id')
        if id != None:
            sql = f"DELETE FROM [{tbl}] WHERE _id = ?"
            con.cursor().execute(sql, [id])

        # Generate new id if not specified
        if id == None or id == "":
            rec['_id'] = util.gen_id()

        rec_cols = list(rec.keys())      # ['_id', '_body', 'field1', 'field2']
        qs = ['?'] * len(rec_cols)       # ['?', '?', '?', '?']

        sql_cols = ", ".join(rec_cols)   # Ex. "_id, _body, field1, field2"
        q_cols   = ", ".join(qs)         # Ex. "?, ?, ?, ?"
        sql = f"INSERT INTO [{tbl}] ({sql_cols}) VALUES ({q_cols})"

        vals = []
        for k in rec_cols:
            vals.append(rec[k])
        con.cursor().execute(sql, vals)
        return rec['_id']

    # Optimistic db update. Run the update first to see if it goes through.
    # If db error, make sure table and cols exist, then retry.
    try:
        recid = _do_update_rec()
    except sqlite3.Error as e:
        # Create table if it doesn't exist.
        init_table(con, tbl)

        # Create new table columns if necessary.
        _add_table_cols(con, tbl, set(rec.keys()))

        recid = _do_update_rec()

    return recid


def _table_cols_sorted(con, tbl):
    """ Return list containing the table's columns with base fields first followed by the
        rest of the fields in alphabetical order.
        Ex. ['_id', '_body', 'amt', 'author', 'footnote]"""
    tbl = _valid_table_name(tbl)

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
    tbl = _valid_table_name(tbl)

    if not tbl in all_tables(con):
        return None

    # select cols always starts with '_id', '_body' followed by the rest of the fields
    # in alphabetical order
    # Ex. "_id, _body, amt, author, footnote"
    scols = ", ".join(_table_cols_sorted(con, tbl))

    sql = f"SELECT {scols} FROM [{tbl}] WHERE _id = ?"
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
    tbl = _valid_table_name(tbl)

    if qwhere == "":
        qwhere = "1"
    if qorder == "":
        qorder = "_id desc"

    if not tbl in all_tables(con):
        return []

    # select cols always starts with '_id', '_body' followed by the rest of the fields
    # in alphabetical order
    # Ex. "_id, _body, amt, author, footnote"
    scols = ", ".join(_table_cols_sorted(con, tbl))

    if q != "":
        sql = f"SELECT {scols} FROM [{tbl}] AS d INNER JOIN [{tbl}_fts] AS fts ON d._id = fts.__id WHERE [{tbl}_fts] MATCH '{q}' AND ({qwhere}) ORDER BY {qorder}"
    else:
        sql = f"SELECT {scols} FROM [{tbl}] WHERE {qwhere} ORDER BY {qorder}"
    cur = con.cursor()

    try:
        cur.execute(sql)
    except Exception as e:
        raise Exception(f"{e}:\n\n{sql}")

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
    tbl = _valid_table_name(tbl)

    if not tbl in all_tables(con):
        return None

    sql = f"DELETE FROM [{tbl}] WHERE _id = ?"
    con.cursor().execute(sql, [id])


def main():
    con = connect_db("test/test.db")

    # Use a table name with non alphanumeric chars and spaces to verify that any
    # table name will work.
    # The only restricted table name chars are []() since they are used for quoting.
    #tbl = """@#-!? abc 123 _:;"' ~`<+-}(=>[*){]"""
    tbl = "robtest"
    purge_table(con, tbl)

    mktext1 = """
.amt 120.00
.cat commute
.dt @@today

Taxi expense
"""

    mktext2 = """
.amt 165.50
.cat dine_out
.dt @@today

Food court
"""

    mktext3 = """
.amt 789.25
.cat utilities
.due_date 2019-04-15
.dt 2019-04-01

Electric bill
"""

    mktext4 = """
.cat todo
.empty_field
.empty_field_with_spaces       
.amt 1
._field_with_underscore This is invalid and will be skipped



Rec with empty or invalid fields.

And extra lines at the start and end that will be stripped out.


"""

    ids = []
    for mktext in [mktext1, mktext2, mktext3, mktext4]:
    #for mktext in [mktext4]:
        print("Updating rec...")
        rec = rec_from_mktext(mktext)
        print(rec)

        new_id = update_rec(con, tbl, rec)
        ids.append(new_id)
        print(f"New rec id: {new_id}")

    for id in ids:
        print(f"Reading rec id: {id}...")
        rec = read_rec(con, tbl, id)
        mktext = mktext_from_rec(rec)
        print(mktext)

    q = "bill"
    qwhere = "amt > 100.0 and due_date >= '2019-04-10'"
    print(f"Reading recs where q='{q}', qwhere='{qwhere}'")
    recs = read_recs(con, tbl, q, qwhere)
    for rec in recs:
        mktext = mktext_from_rec(rec)
        print(mktext)

        # Add a new field 'footnote' and change 'amt' field.
        rec["footnote"] = "This record was edited."
        rec["amt"] += 10
        mktext = mktext_from_rec(rec)
        update_rec(con, tbl, rec)

        print("Rec was updated to:")
        rec = read_rec(con, tbl, rec["_id"])
        mktext = mktext_from_rec(rec)
        print(mktext)

#    for id in ids:
#        print(f"Deleting rec '{id}'...")
#        del_rec(con, tbl, id)

    commit(con)


if __name__ == "__main__":
    main()

