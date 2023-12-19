-- Create enum type "notification_target_type"
CREATE TYPE "public"."notification_target_type" AS ENUM ('discord', 'email');
-- Create "notification_targets" table
CREATE TABLE "public"."notification_targets" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "type" "public"."notification_target_type" NOT NULL, "destination" text NOT NULL, PRIMARY KEY ("id"));
-- Create index "unique_destination" to table: "notification_targets"
CREATE UNIQUE INDEX "unique_destination" ON "public"."notification_targets" ("destination");
