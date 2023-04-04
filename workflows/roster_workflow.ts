import {DefineWorkflow, Schema} from "deno-slack-sdk/mod.ts";
import {RosterMessageFunction} from "../functions/message_function.ts";

const RosterWorkflow = DefineWorkflow({
    callback_id: "roster_workflow",
    title: "Roster workflow",
    description: "Workflow to view a team roster",
    input_parameters: {
        properties: {
            interactivity: {
                type: Schema.slack.types.interactivity,
            },
            channel: {
                type: Schema.slack.types.channel_id,
            },
        },
        required: ["interactivity"],
    },
});

const formData = RosterWorkflow.addStep(
    Schema.slack.functions.OpenForm,
    {
        title: "Recruit a team member",
        interactivity: RosterWorkflow.inputs.interactivity,
        submit_label: "Recruit",
        description: "Recrauit a team member",
        fields: {
            elements: [
                {
                    name: "team",
                    title: "Team",
                    type: Schema.types.string,
                },
                {
                    name: "channel",
                    title: "Channel",
                    type: Schema.slack.types.channel_id,
                },
            ],
            required: ["team", "channel"],
        },
    },
);

RosterWorkflow.addStep(RosterMessageFunction, {
    interactivity: formData.outputs.interactivity,
    channel: formData.outputs.fields.channel,
    team: formData.outputs.fields.team,
})

export default RosterWorkflow;
