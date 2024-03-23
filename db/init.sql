BEGIN;

-- postgresql tables
CREATE TABLE shelves (
       id serial PRIMARY KEY,
       title text NOT NULL
);

CREATE TABLE commodities (
       id serial PRIMARY KEY,
       title text NOT NULL
);

CREATE TABLE commodities_shelves (
       commodity integer REFERENCES commodities(id)
             ON DELETE CASCADE,
       shelf integer REFERENCES shelves(id)
             ON DELETE CASCADE,
       is_main_shelf boolean NOT NULL,
       PRIMARY KEY (commodity, shelf)
);

-- This table doesn't have unique order ids.
-- | id | commodity           | quantity |
-- |----+---------------------+----------|
-- |  1 | 1 (ref. laptop)     |        2 |
-- |  1 | 2 (ref. smartphone) |        1 |
-- |  2 | 1 (ref. laptop)     |        1 |
-- etc. etc.

-- Order '1' has 2 laptops and 1 smartphone, order '2' has
-- 1 laptop, etc.
-- However the orders do autoincrement on when using INSERT
CREATE TABLE orders (
       id serial,
       commodity integer REFERENCES commodities(id)
                 ON DELETE CASCADE,
       quantity integer NOT NULL DEFAULT 1,
       PRIMARY KEY (id, commodity)
);

COMMIT;
