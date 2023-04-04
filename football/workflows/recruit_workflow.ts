import { DefineWorkflow, Schema } from "deno-slack-sdk/mod.ts";

const RecruitWorkflow = DefineWorkflow({
  callback_id: "recruit_workflow",
  title: "Recruit workflow",
  description: "Workflow to recruit a team member",
  input_parameters: {},
});

const recruitForm = RecruitWorkflow.addStep(
    Schema.slack.functions.OpenForm,
    {
      title: "Recruit a team member",
      submit_label: "Recruit",
      fields: {
        elements: [{
          name: "team",
          title: "Team",
          type: Schema.slack.types.usergroup_id,
        }, {
          name: "member",
          title: "Member",
          type: Schema.slack.types.user_id,
        }],
        required: ["team", "member"],
      },
    },
);

export default RecruitWorkflow;
