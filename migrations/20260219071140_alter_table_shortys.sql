-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
DROP INDEX IF EXISTS idx_shortys_original_url;
CREATE UNIQUE INDEX idx_shortys_original_url ON shortys(original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_shortys_original_url;
CREATE INDEX IF NOT EXISTS idx_shortys_original_url ON shortys(original_url);
-- +goose StatementEnd
