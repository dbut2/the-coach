import { Trigger } from "deno-slack-api/types.ts";
import RosterWorkflow from "../workflows/roster_workflow.ts";
/**
 * Triggers determine when workflows are executed. A trigger
 * file describes a scenario in which a workflow should be run,
 * such as a user pressing a button or when a specific event occurs.
 * https://api.slack.com/future/triggers
 */
const rosterTrigger: Trigger<typeof RosterWorkflow.definition> = {
  type: "shortcut",
  name: "Roster trigger",
  description: "Trigger roster workflow",
  workflow: "#/workflows/roster_workflow",
  inputs: {
    interactivity: {
      value: "{{data.interactivity}}",
    },
  },
};

export default rosterTrigger;
