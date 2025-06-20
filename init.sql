CREATE TABLE IF NOT EXISTS users (
    name TEXT PRIMARY KEY,
    passHash TEXT,
    admin BOOL
);

CREATE TABLE IF NOT EXISTS posts ( 
    id SERIAL PRIMARY KEY, 
    author TEXT, 
    postText TEXT, 
    postTime TIMESTAMP, 
    data TEXT, 
    parentId INTEGER, 
    board TEXT
);