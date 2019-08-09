import re
import sqlite3

import util
import db

def main():
    con = db.connect_db()
    tbl = "art_of_worldly_wisdom"

    db.purge_table(con, tbl)
    db.init_table(con, tbl)

    cur = con.cursor()
    sql = f"INSERT INTO {tbl} (_id, _body) VALUES (?, ?)"

    body_lines = []

    is_last_line_page_break = False

    with open("aww.txt", "r") as aww:
        for line in aww:
            line = line.rstrip()

            # Skip over "[p. nnn]" lines
            if re.search(r'^\[p\..*]', line):
                is_last_line_page_break = True
                continue

            # Skip the line after the "[p. nnn]" line
            if is_last_line_page_break and line.strip() == "":
                is_last_line_page_break = False
                continue

            # Title: "iii Keep Matters for a Time in Suspense."
            if re.search(r'^[ivxlcdm]+\s+\w+', line):
                body = "\n".join(body_lines).strip()
                print(body)
                print("---")
                cur.execute(sql, [util.gen_id(), body])

                body_lines = []

            line = line.replace("[paragraph continues] ", "")
            body_lines.append(line)

        if len(body_lines) > 0:
            body = "\n".join(body_lines).strip()
            print(body)
            print("\n---\n")
            cur.execute(sql, [util.gen_id(), body])

    con.commit()


if __name__ == "__main__":
    main()


