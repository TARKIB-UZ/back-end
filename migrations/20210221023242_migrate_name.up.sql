CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    first_name TEXT,
    last_name TEXT,
    phone_number TEXT,
    nickname TEXT,
    password TEXT,
    avatar TEXT
)