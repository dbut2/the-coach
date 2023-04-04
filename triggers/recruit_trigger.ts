import { Trigger } from "deno-slack-api/types.ts";
import RecruitWorkflow from "../workflows/recruit_workflow.ts";
/**
 * Triggers determine when workflows are executed. A trigger
 * file describes a scenario in which a workflow should be run,
 * such as a user pressing a button or when a specific event occurs.
 * https://api.slack.com/future/triggers
 */
const recruitTrigger: Trigger<typeof RecruitWorkflow.definition> = {
  type: "shortcut",
  name: "Recruit trigger",
  description: "Trigger recruit workflow",
  workflow: "#/workflows/recruit_workflow",
  inputs: {
    interactivity: {
      value: "{{data.interactivity}}",
    },
  },
};

export default recruitTrigger;
