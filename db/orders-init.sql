BEGIN;
    INSERT INTO orders (id, commodity, quantity)
           VALUES
           (10, 1, 2),
           (11, 2, 3),
           (14, 1, 3),
           (10, 3, 1),
           (14, 4, 4),
           (15, 5, 1),
           (10, 6, 1);
END;
