-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    Movie (
        ID VARCHAR(255) PRIMARY KEY,
        Title VARCHAR(255),
        Description TEXT,
        Director VARCHAR(255)
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop TABLE IF EXISTS Movie;

-- +goose StatementEnd