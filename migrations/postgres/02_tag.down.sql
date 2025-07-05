ALTER TABLE questions ADD COLUMN tag TEXT NOT NULL DEFAULT '';

UPDATE questions q
SET tag = t.tag
FROM tags t
WHERE q.tag_id = t.id AND q.deleted_at IS NULL;

ALTER TABLE questions DROP CONSTRAINT fk_questions_tag;

ALTER TABLE questions DROP COLUMN tag_id;

DROP TABLE tags;
