BEGIN;

DROP INDEX balls@unique_brand_name_approved_at CASCADE;

ALTER TABLE balls
ADD CONSTRAINT unique_brand_name
UNIQUE (brand, name);

COMMIT;
