CREATE TABLE IF NOT EXISTS action_units (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    muscle TEXT NOT NULL,
    clues TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS questions (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS question_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_id TEXT NOT NULL,
    side TEXT,
    code TEXT NOT NULL,
    intensity TEXT,
    FOREIGN KEY(question_id) REFERENCES questions(id) ON DELETE CASCADE
);

-- 主要AUの初期マスターデータ
INSERT OR IGNORE INTO action_units (code, name, muscle, clues) VALUES
('1', 'Inner Brow Raiser', '前頭筋 内側部', '眉の内側が持ち上がり額中央に水平ジワ'),
('2', 'Outer Brow Raiser', '前頭筋 外側部', '眉の外側が持ち上がる'),
('4', 'Brow Lowerer', '皺眉筋/鼻根筋', '眉が下がり眉間に縦ジワが寄る'),
('5', 'Upper Lid Raiser', '上眼瞼挙筋', '上まぶたが上がり白目の上部が露出'),
('6', 'Cheek Raiser', '眼輪筋 眼窩部', '頬が持ち上がり目尻にカラスの足跡'),
('7', 'Lid Tightener', '眼輪筋 眼瞼部', '下まぶたが緊張して細目になる'),
('9', 'Nose Wrinkler', '上唇鼻翼挙筋', '鼻背にシワが寄り上唇中央が引き上がる'),
('10', 'Upper Lip Raiser', '上唇挙筋', '上唇全体が引き上がり前歯が見える'),
('12', 'Lip Corner Puller', '大頬骨筋', '口角が斜め上に引っ張られる(笑顔)'),
('14', 'Dimpler', '頬筋', '口角が外側奥へ締め込まれエクボができる'),
('15', 'Lip Corner Depressor', '口角下制筋', '口角が下がる(への字口)'),
('17', 'Chin Raiser', 'オトガイ筋', '下唇が押し上げられ顎に梅干しジワ'),
('20', 'Lip Stretcher', '笑筋', '口角が水平外側に引っ張られる'),
('23', 'Lip Tightener', '口輪筋', '唇が薄く引き締まる'),
('25', 'Lips Part', '下顎下制', '唇が開き歯や口腔が見える'),
('26', 'Jaw Drop', '咬筋弛緩', '顎が脱力して自然に落ちる');
