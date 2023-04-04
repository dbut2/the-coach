import { DefineDatastore, Schema } from "deno-slack-sdk/mod.ts";

const TeamsDatastore = DefineDatastore({
  name: "teams",
  primary_key: "team_id",
  attributes: {
    team_id: {
      type: Schema.slack.types.usergroup_id,
    },
    team_members: {
      type: Schema.types.array,
      items: {
        type: Schema.slack.types.user_id,
      }
    },
  },
});

export default TeamsDatastore;
