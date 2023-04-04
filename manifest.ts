import {Manifest} from "deno-slack-sdk/mod.ts";
import TeamsDatastore from "./datastores/teams_datastore.ts";
import RecruitWorkflow from "./workflows/recruit_workflow.ts";
import {RecruitFunction} from "./functions/recruit_function.ts";
import BootWorkflow from "./workflows/boot_workflow.ts";
import {BootFunction} from "./functions/boot_function.ts";
import {RosterMessageFunction} from "./functions/message_function.ts";
import RosterWorkflow from "./workflows/roster_workflow.ts";

/**
 * The app manifest contains the app's configuration. This
 * file defines attributes like app name and description.
 * https://api.slack.com/future/manifest
 */
export default Manifest({
    name: "football",
    description: "A template for building Slack apps with Deno",
    icon: "assets/default_new_app_icon.png",
    workflows: [RosterWorkflow, RecruitWorkflow, BootWorkflow],
    functions: [RecruitFunction, BootFunction, RosterMessageFunction],
    datastores: [TeamsDatastore],
    outgoingDomains: [],
    botScopes: [
        "commands",
        "chat:write",
        "chat:write.public",
        "datastore:read",
        "datastore:write",
    ],
});
