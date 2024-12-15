-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    IF NOT EXISTS ratings (
        record_id VARCHAR(255),
        record_type VARCHAR(255),
        user_id VARCHAR(255),
        value INT
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop TABLE IF EXISTS ratings;

-- +goose StatementEnd