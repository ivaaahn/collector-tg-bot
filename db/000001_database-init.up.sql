CREATE TABLE "users"
(
    "id"         bigint PRIMARY KEY,
    "username"   text,
    "created_at" timestamptz
);

CREATE TABLE "sessions"
(
    "id"          uuid PRIMARY KEY,
    "title"       text,
    "creator_id"  bigint,
    "chat_id"     bigint,
    "created_at"  timestamptz,
    "finished_at" timestamptz null default null
);

CREATE TABLE "members"
(
    "user_id"    bigint,
    "session_id" uuid,
    PRIMARY KEY ("user_id", "session_id")
);

CREATE TABLE "purchases"
(
    "id"         serial primary key,
    "buyer_id"   bigint,
    "session_id" uuid,
    "price"      int,
    "created_at" timestamptz default now(),
    "title"      text,
    "quantity"   int
);

CREATE TABLE "expenses"
(
    "purchase_id" int,
    "eater_id"    bigint,
    "session_id"  uuid,
    "qty"         int DEFAULT 1,
    PRIMARY KEY ("purchase_id", "eater_id", "session_id")
);

CREATE UNIQUE INDEX ON "users" ("username");

ALTER TABLE "sessions"
    ADD FOREIGN KEY ("creator_id") REFERENCES "users" ("id");

ALTER TABLE "members"
    ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id"),
    ADD FOREIGN KEY ("session_id") REFERENCES "sessions" ("id");

ALTER TABLE "purchases"
    ADD FOREIGN KEY ("buyer_id") REFERENCES "users" ("id"),
    ADD FOREIGN KEY ("session_id") REFERENCES "sessions" ("id"),
    ADD FOREIGN KEY ("session_id", "buyer_id") REFERENCES "members" ("session_id", "user_id");

ALTER TABLE "expenses"
    ADD FOREIGN KEY ("purchase_id") REFERENCES "purchases" ("id"),
    ADD FOREIGN KEY ("eater_id") REFERENCES "users" ("id"),
    ADD FOREIGN KEY ("session_id") REFERENCES "sessions" ("id"),
    ADD FOREIGN KEY ("eater_id", "session_id") REFERENCES "members" ("user_id", "session_id");
