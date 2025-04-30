BEGIN;

DROP INDEX balls@unique_brand_name CASCADE;

ALTER TABLE balls
ADD CONSTRAINT unique_brand_name_approved_at
UNIQUE (brand, name, approved_at);

COMMIT;
