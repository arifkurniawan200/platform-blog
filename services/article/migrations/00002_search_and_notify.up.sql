-- Add full-text search index on articles
CREATE INDEX IF NOT EXISTS idx_articles_fts
  ON articles
  USING GIN (to_tsvector('english', title || ' ' || COALESCE(subtitle, '') || ' ' || content));
