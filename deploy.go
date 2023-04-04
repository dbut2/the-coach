package football

//go:generate gcloud functions deploy pass-ball --project dbut-0 --region australia-southeast1 --allow-unauthenticated --entry-point PassBall --runtime go120 --set-secrets SLACK_SIGNING_SECRET=projects/574604089887/secrets/passball-slack-signing-secret/versions/latest,SLACK_BOT_TOKEN=projects/574604089887/secrets/passball-slack-bot-token/versions/latest --trigger-http
