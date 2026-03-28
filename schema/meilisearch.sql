-- Meilisearch 索引文档字段约定（与同步到索引的 JSON 字段对齐）
index medias (
    id VARCHAR(64) PRIMARY KEY,
    key VARCHAR(255),
    title VARCHAR(255),
    keywords VARCHAR(255),
    content TEXT,
    summary VARCHAR(255),
    remark VARCHAR(255),
    media_id VARCHAR(16),
    media_name VARCHAR(64),
    categories_ids VARCHAR(255),
    categories_names VARCHAR(255),
    tags_ids VARCHAR(255),
    tags_names VARCHAR(255),
    status INT
);
