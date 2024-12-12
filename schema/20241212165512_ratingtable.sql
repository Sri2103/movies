-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    Rating (
        ID VARCHAR(255) PRIMARY KEY,
        UserID VARCHAR(255),
        MovieID VARCHAR(255),
        Value INT
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop TABLE IF EXISTS rating;

-- +goose StatementEnd