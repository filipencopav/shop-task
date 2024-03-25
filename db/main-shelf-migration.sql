BEGIN;
ALTER TABLE commodities
ADD COLUMN main_shelf integer;

UPDATE commodities AS comms
    SET main_shelf = (SELECT shelf
                      FROM commodities_shelves
                      WHERE is_main_shelf AND commodity = comms.id);

ALTER TABLE commodities_shelves DROP COLUMN is_main_shelf;
END;
