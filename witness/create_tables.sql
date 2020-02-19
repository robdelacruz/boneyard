PRAGMA foreign_keys = ON;

DROP TABLE IF EXISTS log;
DROP TABLE IF EXISTS topic;

CREATE TABLE IF NOT EXISTS post (
    post_id INTEGER NOT NULL PRIMARY KEY,
    post_type INTEGER DEFAULT 0,             -- (a)
    post_createdt TEXT DEFAULT (date('now')),
    post_body TEXT DEFAULT '',
    post_startdt TEXT NULL,                  -- (b)
    post_enddt TEXT NULL,                    -- (b)
    post_status INTEGER DEFAULT 0,           -- (c)
    topic_id INTEGER DEFAULT 0,
    FOREIGN KEY(topic_id) REFERENCES topic
);

/* (a)
post_type
------------
What type of post this is.

0: log
1: note
2: event
3: task
*/

/* (b)
post_startdt, post_enddt
------------------------------
Only applies to 'event' and 'task' post_type.

For 'event', refers to the start and end dates of the event.
For 'task', post_startdt refers to the 'alert' date,
            post_enddt refers to the 'due' date.
*/

/* (c)
post_status
--------------
Only applies to 'task' post_type.
Determines whether the task has been completed or not.

0: not completed
1: completed
*/

CREATE TABLE IF NOT EXISTS topic (
    topic_id INTEGER NOT NULL PRIMARY KEY,
    topic_name TEXT NOT NULL
);
INSERT INTO topic (topic_id, topic_name) VALUES (0, "");

