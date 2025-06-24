CREATE TABLE IF NOT EXISTS users (
    name TEXT PRIMARY KEY,
    passHash TEXT,
    admin BOOL
);

CREATE TABLE IF NOT EXISTS posts ( 
    id SERIAL PRIMARY KEY, 
    author TEXT, 
    postText TEXT, 
    postTime INT, 
    data TEXT, 
    parentId INTEGER, 
    board TEXT,
    responses INT,
    lastAnswer INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS boards (
    name TEXT PRIMARY KEY,
    description TEXT,
    admins TEXT[],
    mode INT,
    usersList TEXT[],
    owner TEXT,
    posts INT,
    postsWithR INT
);