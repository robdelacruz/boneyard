import sqlite3

dbfile = "nbdata.db"
con = sqlite3.connect(dbfile)
con.row_factory = sqlite3.Row
cur = con.cursor()

sql = """CREATE TABLE IF NOT EXISTS nbdata (
  id TEXT,
  title TEXT,
  body TEXT
);"""
cur.execute(sql)
con.commit()

sql = """INSERT INTO nbdata (id, title, body) 
VALUES(?, ?, ?)"""
#cur.execute(sql, ("abc", "abc title", ""))
#except sqlite3.Error as e:
#con.commit()

sql = """SELECT id, title, body FROM nbdata ORDER BY id"""
cur.execute(sql)
for row in cur:
  #print(f"{row[0]}: --{row[1]}--")
  print(row.keys())
  id = row["id"]
  title = row["title"]
  print("%s: --%s--" % (id, title))

con.close()

