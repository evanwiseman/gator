-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY, -- UUID 
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT,
    url TEXT UNIQUE,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID,
    CONSTRAINT fk_feed_id
        FOREIGN KEY (feed_id)
        REFERENCES feeds(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;