import { DefineFunction, Schema, SlackFunction } from "deno-slack-sdk/mod.ts";
import SampleObjectDatastore from "../datastores/teams_datastore.ts";
import TeamsDatastore from "../datastores/teams_datastore.ts";

export const RecruitFunctionDefinition = DefineFunction({
  callback_id: "recruit_function",
  title: "Recruit function",
  description: "Recruit a team member",
  source_file: "functions/recruit_function.ts",
  input_parameters: {
    properties: {
      team_id: {
        type: Schema.slack.types.usergroup_id,
        description: "Team to recruit to",
      },
      team_member: {
        type: Schema.slack.types.user_id,
        description: "The user to recruit",
      },
    },
    required: ["team_id", "team_member"],
  },
});

export default SlackFunction(
  RecruitFunctionDefinition,
  async ({ inputs, client }) => {
    const getResponse = await client.apps.datastore.get<typeof TeamsDatastore.definition>({
        datastore: "teams",
        id: inputs.team,
    });

    if (!getResponse.ok) {
        const error = `Failed to get team from datastore: ${getResponse.error}`;
        return { error };
    }

    if (getResponse.item.members.includes(inputs.member)) {
        const error = `Member has already been recruited`
        return { error };
    }

    const members = getResponse.item.members;

    members.push(inputs.member);

    const updateResponse = await client.apps.datastore.update<typeof TeamsDatastore.definition>({
        datastore: "teams",
        item: {
            team_id: inputs.team,
            members: members,
        }
    });

    if (!updateResponse.ok) {
        const error = `Failed to update team to datastore: ${updateResponse.error}`;
        return { error };
    }

    return { outputs: {} };
  },
);
