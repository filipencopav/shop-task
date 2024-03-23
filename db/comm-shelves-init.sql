BEGIN;
    -- А
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (1, 1, true);
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (2, 1, true);
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (5, 1, false);

    -- Б
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (3, 2, true);

    -- В
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (3, 3, false);

    -- Ж
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (4, 4, true);
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (5, 4, true);
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (6, 4, true);

    -- З
    INSERT INTO commodities_shelves (commodity, shelf, is_main_shelf)
           VALUES (3, 5, false);
END;
