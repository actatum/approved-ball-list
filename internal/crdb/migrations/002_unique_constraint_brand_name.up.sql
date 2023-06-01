BEGIN;

-- CREATE UNIQUE INDEX CONCURRENTLY brand_name
-- ON balls(brand, name);

ALTER TABLE balls
ADD CONSTRAINT unique_brand_name
UNIQUE (brand, name);

COMMIT;