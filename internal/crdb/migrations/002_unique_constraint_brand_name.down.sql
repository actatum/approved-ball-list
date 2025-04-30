BEGIN;

ALTER TABLE balls
DROP CONSTRAINT unique_brand_name;

COMMIT;
