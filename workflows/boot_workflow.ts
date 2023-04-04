import {DefineWorkflow, Schema} from "deno-slack-sdk/mod.ts";
import {RecruitFunction} from "../functions/recruit_function.ts";

const BootWorkflow = DefineWorkflow({
    callback_id: "boot_workflow",
    title: "Boot workflow",
    description: "Workflow to boot a team member",
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

const formData = BootWorkflow.addStep(
    Schema.slack.functions.OpenForm,
    {
        title: "Recruit a team member",
        interactivity: BootWorkflow.inputs.interactivity,
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
                    name: "member",
                    title: "Member",
                    type: Schema.slack.types.user_id,
                },
            ],
            required: ["team", "member"],
        },
    },
);

BootWorkflow.addStep(RecruitFunction, {
    interactivity: formData.outputs.interactivity,
    team: formData.outputs.fields.team,
    member: formData.outputs.fields.member,
});

export default BootWorkflow;
