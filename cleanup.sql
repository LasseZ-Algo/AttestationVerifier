DROP TABLE IF EXISTS reports;
CREATE TABLE reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_hash TEXT UNIQUE,
    report TEXT,
    verified INTEGER
);