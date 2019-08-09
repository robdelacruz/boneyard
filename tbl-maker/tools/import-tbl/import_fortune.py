import re
import sqlite3
import sys
import os

import util
import db

def main():
    if len(sys.argv) < 2:
        print("Usage: init_fortune.py <fortune_filename> [table name]")
        sys.exit(1)

    fortune_file = sys.argv[1]

    # If [table name] specified, use it. Otherwise get table name from fortune filename.
    if len(sys.argv) >= 3:
        tbl = sys.argv[2]
    else:
        (sdir, filename) = os.path.split(fortune_file)
        tbl = os.path.splitext(filename)[0]

    con = db.connect_db()

    db.purge_table(con, tbl)
    db.init_table(con, tbl)

    cur = con.cursor()
    sql = f"INSERT INTO {tbl} (_id, _body) VALUES (?, ?)"

    body_lines = []

    with open(fortune_file, "r") as f:
        # Skip over lines until the first "%" line
        #for line in f:
        #    line = line.rstrip()
        #    print(line)
        #    if re.search(r'^%\s*$', line):
        #        break

        for line in f:
            line = line.rstrip()

            # "%" separator line
            if re.search(r'^%\s*$', line):
                body = "\n".join(body_lines).strip()
                if body != "":
                    print(body)
                    print("---")
                    cur.execute(sql, [util.gen_id(), body])

                body_lines = []
                continue

            body_lines.append(line)

        body = "\n".join(body_lines).strip()
        if body != "":
            print(body)
            print("---")
            cur.execute(sql, [util.gen_id(), body])

    con.commit()


if __name__ == "__main__":
    main()


