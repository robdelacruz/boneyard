DROP TABLE IF EXISTS nbdata_fts;
CREATE VIRTUAL TABLE IF NOT EXISTS nbdata_fts USING fts5(_id, _title, _body, tokenize = porter);

DROP TRIGGER IF EXISTS nbdata_insert;
DROP TRIGGER IF EXISTS nbdata_update;
DROP TRIGGER IF EXISTS nbdata_delete;

CREATE TRIGGER IF NOT EXISTS nbdata_insert AFTER INSERT ON nbdata BEGIN
  INSERT INTO nbdata_fts (_id, _title, _body)
  VALUES (new.id, new.title, new.body);
END;

CREATE TRIGGER IF NOT EXISTS nbdata_update AFTER UPDATE OF title, body ON nbdata BEGIN
  UPDATE nbdata_fts
  SET _title = new.title, _body = new.body
  WHERE _id = old.id;
END;

CREATE TRIGGER IF NOT EXISTS nbdata_delete AFTER DELETE ON nbdata BEGIN
  DELETE FROM nbdata_fts WHERE _id = old.id;
END;

INSERT INTO nbdata_fts (_id, _title, _body) SELECT id, title, body FROM nbdata;

