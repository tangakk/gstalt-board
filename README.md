Нужна постгрес бд:

CREATE TABLE posts (
id SERIAL PRIMARY KEY,
author TEXT,
postText TEXT,
postTime TIMESTAMP,
data TEXT,
parentId INTEGER,
board TEXT);


