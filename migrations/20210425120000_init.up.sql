CREATE TABLE IF NOT EXISTS developers (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    email TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS packages (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  md5sum TEXT NOT NULL,
  date_publication TIMESTAMP WITH TIME ZONE,
  title TEXT,
  description TEXT,
  author_id INT NOT NULL,
  maintainer_id INT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT packages_author_id
        FOREIGN KEY(author_id)
      REFERENCES developers(id),
  CONSTRAINT packages_maintainer_id
        FOREIGN KEY(maintainer_id)
      REFERENCES developers(id),
  UNIQUE (name, version)
);
