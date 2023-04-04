import { Trigger } from "deno-slack-api/types.ts";
import BootWorkflow from "../workflows/boot_workflow.ts";
/**
 * Triggers determine when workflows are executed. A trigger
 * file describes a scenario in which a workflow should be run,
 * such as a user pressing a button or when a specific event occurs.
 * https://api.slack.com/future/triggers
 */
const bootTrigger: Trigger<typeof BootWorkflow.definition> = {
  type: "shortcut",
  name: "Boot trigger",
  description: "Trigger recruit workflow",
  workflow: "#/workflows/recruit_workflow",
  inputs: {
    interactivity: {
      value: "{{data.interactivity}}",
    },
  },
};

export default bootTrigger;
