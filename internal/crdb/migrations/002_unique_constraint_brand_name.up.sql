BEGIN;

ALTER TABLE balls
ADD CONSTRAINT unique_brand_name
UNIQUE (brand, name);

COMMIT;
