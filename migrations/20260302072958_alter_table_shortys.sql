-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE shortys ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_shortys_user_id ON shortys(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_shortys_user_id;
ALTER TABLE shortys DROP COLUMN is_deleted;
-- +goose StatementEnd
