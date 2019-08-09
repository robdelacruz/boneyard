#! /bin/sh

p1=\"$@\"
sql="SELECT id, title, body FROM nbdata AS d INNER JOIN nbdata_fts AS fts ON id = fts._id WHERE nbdata_fts MATCH '$p1' AND body like '%%' ORDER BY fts.rank;"

#echo sql: $sql
echo $sql | sqlite3 nbdata.db

