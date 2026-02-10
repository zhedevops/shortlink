-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE shortys (
                         id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                         short_url VARCHAR(8) NOT NULL,
                         original_url TEXT NOT NULL,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_shortys_short_url ON shortys(short_url);
CREATE INDEX idx_shortys_original_url ON shortys(original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_shortys_original_url;
DROP INDEX IF EXISTS idx_shortys_short_url;
DROP TABLE IF EXISTS shortys CASCADE;
-- +goose StatementEnd
