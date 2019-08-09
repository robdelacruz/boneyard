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
        # Use ~/.grouptest/data.db by default if no file specified.
        dbdir = os.path.join(str(pathlib.Path.home()), ".grouptest")
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


def main():
    con = connect_db("test/test.db")


if __name__ == "__main__":
    main()

