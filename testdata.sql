INSERT INTO songs (name, group_name, text, release_date, link)
SELECT 
    'Song ' || i,
    'Group ' || (i % 10 + 1),
    'Lyrics for song ' || i,
    CURRENT_DATE - (i % 365),
    'https://example.com'
FROM generate_series(1, 100) AS s(i);
