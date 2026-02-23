-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE shortys ADD COLUMN user_id INT NOT NULL, ADD CONSTRAINT fk_user_id_shortys FOREIGN KEY (user_id) REFERENCES users (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE shortys DROP CONSTRAINT fk_user_id_shortys;
ALTER TABLE shortys DROP COLUMN user_id;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
