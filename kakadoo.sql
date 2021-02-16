CREATE TABLE games (
    id serial NOT NULL CONSTRAINT games_pk PRIMARY KEY,
    title varchar(128) NOT NULL,
    author varchar(32) NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);
CREATE TABLE quiz (
    id serial NOT NULL CONSTRAINT quiz_pk PRIMARY KEY,
    game_id integer NOT NULL CONSTRAINT quiz_games_id_fk REFERENCES games ON DELETE CASCADE,
    question varchar(255) NOT NULL,
    answers varchar(255) [] NOT NULL,
    correct_answer_idx integer NOT NULL,
    time_to_answer integer DEFAULT 5 NOT NULL,
    created_at timestamp DEFAULT current_timestamp NOT NULL
);
