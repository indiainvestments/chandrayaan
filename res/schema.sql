CREATE DATABASE cy;

\c cy;

CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION pg_trgm;

CREATE TABLE IF NOT EXISTS bse_scrip
(
    security_code  integer,
    security_id    text unique,
    security_name  text,
    status         text,
    security_group text,
    face_value     double precision,
    isin_number    text primary key,
    industry       text,
    security_type  text
);

CREATE INDEX ON bse_scrip USING GIN (UPPER(security_name) gin_trgm_ops, security_id gin_trgm_ops);

CREATE TABLE IF NOT EXISTS users
(
    id            text PRIMARY KEY,
    notifications bool,
    created_at    timestamp default now(),
    deleted_at    timestamp default NULL,
    updated_at    timestamp default now()
);

CREATE TABLE IF NOT EXISTS user_scrip
(
    user_id    text,
    ticker     text,
    created_at timestamp default now(),
    deleted_at timestamp default NULL,
    updated_at timestamp default now(),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (ticker) REFERENCES bse_scrip (security_id),
    PRIMARY KEY (user_id, ticker)
);


CREATE TABLE IF NOT EXISTS corporate_news
(
    attachment text,
    id         text primary key,
    headline   text,
    date       timestamp,
    category   text,
    ticker     text,
    news_sub   text,
    foreign key (ticker) references bse_scrip (security_id)
);

CREATE TABLE IF NOT EXISTS corporate_action
(
    id           serial primary key,
    ex_date      date,
    purpose      text,
    details      text,
    payment_date date,
    ticker       text,
    foreign key (ticker) references bse_scrip (security_id)
);