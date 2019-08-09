SELECT * FROM nbdata AS d
INNER JOIN nbdata_fts AS fts ON d._id = fts.__id
WHERE nbdata_fts MATCH 'fortune'
ORDER BY d._id;

