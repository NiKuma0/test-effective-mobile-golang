CREATE TABLE songs (
	id SERIAL PRIMARY KEY,
	name TEXT,
	group_name TEXT,
	text TEXT,
	release_date DATE DEFAULT CURRENT_DATE,
	link TEXT
)
