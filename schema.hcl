schema "public" {
}

table "balls" {
  schema = schema.public
  column "id" {
    null    = false
    type    = bigint
    default = sql("nextval('public.ball_ids'::REGCLASS)")
  }
  column "brand" {
    null = false
    type = text
  }
  column "name" {
    null = false
    type = text
  }
  column "image_url" {
    null = false
    type = text
  }
  column "approved_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "unique_brand_name" {
    unique  = true
    columns = [column.brand, column.name]
  }
}

enum "notification_target_type" {
  schema = schema.public
  values = ["discord", "email"]
}

table "notification_targets" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  column "type" {
    null = false
    type = enum.notification_target_type
  }
  column "destination" {
    null = false
    type = text
  }
  primary_key  {
    columns = [column.id]
  }
  index "unique_destination" {
    unique = true
    columns = [column.destination]
  }
}

enum "notification_state" {
  schema = schema.public
  values = ["pending", "complete", "errored"]
}

table "notifications" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "state" {
    null = false
    type = enum.notification_state
  }
  column "ball_id" {
    null = false
    type = bigint
  }
  column "target_id" {
    null = false
    type = uuid
  }
  column "sent_at" {
    null = true
    type = timestamptz
  }
  primary_key  {
    columns = [column.id]
  }
  foreign_key "target_id" {
    columns = [column.target_id]
    ref_columns = [table.notification_targets.column.id]
    on_update = NO_ACTION
    on_delete = NO_ACTION
  }
  foreign_key "ball_id" {
    columns = [column.ball_id]
    ref_columns = [table.balls.column.id]
    on_update = NO_ACTION
    on_delete = NO_ACTION
  }
}

// table "schema_lock" {
//   schema = schema.public
//   column "lock_id" {
//     null = false
//     type = bigint
//   }
//   primary_key {
//     columns = [column.lock_id]
//   }
// }
// table "schema_migrations" {
//   schema = schema.public
//   column "version" {
//     null = false
//     type = bigint
//   }
//   column "dirty" {
//     null = false
//     type = boolean
//   }
//   primary_key {
//     columns = [column.version]
//   }
// }

