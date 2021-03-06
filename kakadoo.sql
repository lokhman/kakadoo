CREATE TYPE game_type AS ENUM ('quiz', 'woc');

CREATE TABLE games (
    id serial NOT NULL CONSTRAINT games_pk PRIMARY KEY,
    type game_type NOT NULL,
    title varchar(128) NOT NULL,
    author varchar(32) NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);

CREATE TABLE tasks (
    id serial NOT NULL CONSTRAINT tasks_pk PRIMARY KEY,
    game_id integer NOT NULL CONSTRAINT tasks_games_id_fk REFERENCES games ON DELETE CASCADE,
    question varchar(255) NOT NULL,
    answers varchar(255) [] DEFAULT '{}' NOT NULL,
    correct_answer varchar(255) NOT NULL,
    time_to_answer integer DEFAULT 10 NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);
