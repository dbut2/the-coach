import {DefineFunction, Schema, SlackFunction} from "deno-slack-sdk/mod.ts";
import TeamsDatastore from "../datastores/teams_datastore.ts";
import {DatastoreItem} from "deno-slack-api/typed-method-types/apps.ts";

export const RecruitFunction = DefineFunction({
    callback_id: "recruit_function",
    title: "Recruit function",
    description: "Recruit a team member",
    source_file: "functions/recruit_function.ts",
    input_parameters: {
        properties: {
            team: {
                type: Schema.slack.types.usergroup_id,
                description: "Team to recruit to",
            },
            member: {
                type: Schema.slack.types.user_id,
                description: "The user to recruit",
            },
        },
        required: ["team", "member"],
    },
});

export default SlackFunction(
    RecruitFunction,
    async ({inputs, client}) => {
        const getResponse = await client.apps.datastore.get<typeof TeamsDatastore.definition>({
            datastore: "teams",
            id: inputs.team,
        });

        if (!getResponse.ok) {
            const error = `Failed to get team from datastore: ${getResponse.error}`;
            return {error};
        }

        const members = getResponse.item.team_members

        if (members.includes(inputs.member)) {
            const error = `Member has already been recruited`
            return {error};
        }

        members.push(inputs.member);

        const newTeam: DatastoreItem<typeof TeamsDatastore.definition> = {
            team_id: inputs.team,
            team_members: members,
        }

        const updateResponse = await client.apps.datastore.update<typeof TeamsDatastore.definition>({
            datastore: "teams",
            item: newTeam,
        });

        if (!updateResponse.ok) {
            const error = `Failed to update team to datastore: ${updateResponse.error}`;
            return {error};
        }

        return {outputs: {}};
    },
);
