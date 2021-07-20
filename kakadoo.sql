CREATE TABLE games (
    id serial NOT NULL CONSTRAINT games_pk PRIMARY KEY,
    title varchar(128) NOT NULL,
    author varchar(32) NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);

CREATE TYPE task_type AS ENUM ('quiz', 'woc');

CREATE TABLE tasks (
    id serial NOT NULL CONSTRAINT tasks_pk PRIMARY KEY,
    game_id integer NOT NULL CONSTRAINT tasks_games_id_fk REFERENCES games ON DELETE CASCADE,
    type task_type NOT NULL,
    question varchar(255) NOT NULL,
    answers varchar(255) [] DEFAULT '{}' NOT NULL,
    correct_answer varchar(255) NOT NULL,
    time_to_answer integer DEFAULT 10 NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);
