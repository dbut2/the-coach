# The Coach

### Randomly reassign single member slack user groups with ease

---

![](image.png)

## Usage

Create Slack app with the following scopes:
- `commands`
- `usergroups:read`
- `usergroups:write`
- `chat:write`

Copy Signing Secret and Bot User OAuth Token

Attach `PassBall` function to a router, passing `SLACK_SIGNING_SECRET`, and `SLACK_BOT_TOKEN` env vars.
Point Slack slash command to address.

Optionally if using Cloud Functions you can use `deploy.go`, just update values to your project.
