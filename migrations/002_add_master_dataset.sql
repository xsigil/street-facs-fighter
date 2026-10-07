-- migrations/002_add_master_dataset.sql

-- 既存スキーマがある場合の列追加
ALTER TABLE questions ADD COLUMN item_id TEXT;
ALTER TABLE questions ADD COLUMN split TEXT;         -- 'practice' または 'example'
ALTER TABLE questions ADD COLUMN media_type TEXT;    -- 'image' または 'video'
ALTER TABLE questions ADD COLUMN raw_score TEXT;     -- 原典表記 (例: "7D+9D+17B")
ALTER TABLE questions ADD COLUMN rationale TEXT;     -- ポール・エクマンらの解説文

CREATE INDEX IF NOT EXISTS idx_questions_split ON questions(split);
CREATE INDEX IF NOT EXISTS idx_questions_item_id ON questions(item_id);
