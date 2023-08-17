-- these are queries that ive found helpful in exploring the tags sqlite database
SELECT COUNT(DISTINCT file) FROM file_tags;
SELECT COUNT(DISTINCT file) AS SUMVALUE, file FROM file_tags GROUP BY file ORDER BY SUMVALUE DESC;
SELECT * FROM files where tags_json = "[]";

-- simple queries
SELECT * FROM file_tags where file = '2023-01-06_185137_01.jpg';
SELECT * FROM files where filename = '2023-01-06_185137_01.jpg';

-- an insert
INSERT OR REPLACE INTO files (filename, path, tags_json) VALUES ('2021-01-02_105646_01.jpg', 'C:\Users\Batman\Desktop\2021\2021-01-02_105646_01.jpg', '["calvin","kendrick","parkview-ave","cameron","snow"]');