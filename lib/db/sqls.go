package db

const create_db = `
    CREATE TABLE 'action' (
        'id' INTEGER PRIMARY KEY AUTOINCREMENT,
        'action' VARCHAR(64) NULL,
        'value' TEXT NULL,
        'last_update' DATE,
        'created_at' DATE
    )
`

const action_insert = "INSERT INTO action(action, last_update, created_at) VALUES ('egs', datetime('now'), datetime('now'))"
