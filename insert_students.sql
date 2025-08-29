SET autocommit = 0;
SET unique_checks = 0;
SET foreign_key_checks = 0;

INSERT INTO students (id, name, tel, study) 
SELECT 
    CONCAT('STU', LPAD(seq, 6, '0')) as id,
    CONCAT('学生', seq) as name,
    CONCAT('138', LPAD(FLOOR(10000000 + RAND() * 90000000), 8, '0')) as tel,
    CASE FLOOR(RAND() * 8)
        WHEN 0 THEN '计算机科学'
        WHEN 1 THEN '数学'
        WHEN 2 THEN '物理'
        WHEN 3 THEN '化学'
        WHEN 4 THEN '生物'
        WHEN 5 THEN '统计学'
        WHEN 6 THEN '工程力学'
        WHEN 7 THEN '英语'
    END as study
FROM (
    SELECT @row_number := @row_number + 1 AS seq
    FROM information_schema.columns c1
    CROSS JOIN information_schema.columns c2
    CROSS JOIN information_schema.columns c3
    CROSS JOIN (SELECT @row_number := 0) r
    LIMIT 100000
) numbers;

COMMIT;
SET autocommit = 1;
SET unique_checks = 1;
SET foreign_key_checks = 1;
