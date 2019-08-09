DROP TABLE IF EXISTS nbdata_fts;
CREATE VIRTUAL TABLE IF NOT EXISTS nbdata_fts USING fts5(__id, __body, tokenize = porter);

DROP TRIGGER IF EXISTS nbdata_insert;
DROP TRIGGER IF EXISTS nbdata_update;
DROP TRIGGER IF EXISTS nbdata_delete;

CREATE TRIGGER IF NOT EXISTS nbdata_insert AFTER INSERT ON nbdata BEGIN
  INSERT INTO nbdata_fts (__id, __body)
  VALUES (new._id, new.body);
END;

CREATE TRIGGER IF NOT EXISTS nbdata_update AFTER UPDATE OF body ON nbdata BEGIN
  UPDATE nbdata_fts
  SET __body = new.body
  WHERE __id = old._id;
END;

CREATE TRIGGER IF NOT EXISTS nbdata_delete AFTER DELETE ON nbdata BEGIN
  DELETE FROM nbdata_fts WHERE __id = old._id;
END;

INSERT INTO nbdata_fts (__id, __body) SELECT __id, body FROM nbdata;

