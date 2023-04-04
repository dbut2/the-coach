import {DefineFunction, Schema, SlackFunction} from "deno-slack-sdk/mod.ts";
import TeamsDatastore from "../datastores/teams_datastore.ts";

export const RosterMessageFunction = DefineFunction({
    callback_id: "roster_message_function",
    title: "Roster message function",
    description: "Send a roster message",
    source_file: "functions/message_function.ts",
    input_parameters: {
        properties: {
            channel: {
                type: Schema.slack.types.channel_id,
                description: "channel to send message",
            },
            team: {
                type: Schema.slack.types.usergroup_id,
                description: "team",
            },
        },
        required: ["channel", "team"],
    },
});

export default SlackFunction(
    RosterMessageFunction,
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

        const msgResponse
            = await client.chat.postMessage({
            channel: inputs.channel,
            blocks: [
                {
                    "type": "section",
                    "text": {
                        "type": "plain_text",
                        "text": "Roster",
                        "emoji": true,
                    }
                },
                members.map((member) => {
                    return {
                        "type": "section",
                        "text": member,
                        "emoji": true,
                    }
                })
            ]
        });

        if (!msgResponse) {
            return {error: msgResponse.error}
        }

        return {}
    },
);
