INSERT INTO songs (id, name, group_name, text, release_date)
SELECT 
    i,
    'Song ' || i,
    'Group ' || (i % 10 + 1),
    'Lyrics for song ' || i,
    CURRENT_DATE - (i % 365)
FROM generate_series(1, 100) AS s(i);
