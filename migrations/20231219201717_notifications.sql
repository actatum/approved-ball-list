-- Create enum type "notification_state"
CREATE TYPE "public"."notification_state" AS ENUM ('pending', 'complete', 'errored');
-- Create "notifications" table
CREATE TABLE "public"."notifications" ("id" uuid NOT NULL, "state" "public"."notification_state" NOT NULL, "ball_id" bigint NOT NULL, "target_id" uuid NOT NULL, "sent_at" timestamptz NULL, PRIMARY KEY ("id"), CONSTRAINT "ball_id" FOREIGN KEY ("ball_id") REFERENCES "public"."balls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "target_id" FOREIGN KEY ("target_id") REFERENCES "public"."notification_targets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
