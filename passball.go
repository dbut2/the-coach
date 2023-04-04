package football

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/slack-go/slack"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var (
	slackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET")
	slackBotToken      = os.Getenv("SLACK_BOT_TOKEN")
)

func init() {
	functions.HTTP("PassBall", PassBallFunction)
}

func PassBallFunction(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
	if hadError(err, w, errMessage) {
		return
	}

	r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
	s, err := slack.SlashCommandParse(r)
	if hadError(err, w, errMessage) {
		return
	}

	err = verifier.Ensure()
	if hadError(err, w, errMessage) {
		return
	}

	handleRequestSlash(w, s)
}

func handleRequestSlash(w http.ResponseWriter, s slack.SlashCommand) {
	switch s.Command {
	case "/passball":
		passBall(w, s)
	default:
		handleError(fmt.Errorf("unknown command: %s", s.Command), w, fmt.Sprintf("Unknown command: %s", s.Command))
	}
}

func passBall(w http.ResponseWriter, s slack.SlashCommand) {
	split := strings.Split(strings.TrimSpace(s.Text), " ")

	if len(split) < 1 {
		handleError(errors.New("no usergroup supplied"), w, "Please supply a user group")
		return
	}

	userGroupID, has := parseUserGroupID(split[0])
	if !has {
		handleError(fmt.Errorf("invalid usergroup: %s", split[0]), w, "Please supply a valid user group")
		return
	}

	if len(split) < 2 {
		handleError(errors.New("no receiver supplied"), w, "Please supply a user or user group receiver")
		return
	}

	sc := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	parsed := false
	var newUserID string

	if receiverUserGroupID, has := parseUserGroupID(split[1]); has {
		userIDs, err := sc.GetUserGroupMembers(receiverUserGroupID)
		if hadError(err, w, errMessage) {
			return
		}

		if len(userIDs) > 0 {
			parsed = true
			newUserID = userIDs[rand.Intn(len(userIDs))]
		}
	}

	if receiverUserID, has := parseUserID(split[1]); has {
		parsed = true
		newUserID = receiverUserID
	}

	if !parsed || newUserID == "" {
		handleError(fmt.Errorf("invalid receiver: %s", split[1]), w, "Please supply a valid user or user group receiver")
		return
	}

	_, _, err := sc.PostMessage(s.ChannelID, slack.MsgOptionText(randomPhrase(s.UserID, newUserID), false))
	if hadError(err, w, errMessage) {
		return
	}

	_, err = sc.UpdateUserGroupMembers(userGroupID, newUserID)
	if hadError(err, w, errMessage) {
		return
	}

	w.Header().Set("content-type", "application/json")
	if hadError(err, w, errMessage) {
		return
	}
}

//go:embed phrases.txt
var phrasesFile string

func randomPhrase(from, to string) string {
	var phrases []string
	for _, phrase := range strings.Split(phrasesFile, "\n") {
		if phrase == "" {
			continue
		}
		phrases = append(phrases, phrase)
	}

	t := template.New("phrases")
	t, err := t.Parse(phrases[rand.Intn(len(phrases))])
	if err != nil {
		panic(err.Error())
	}

	buf := &bytes.Buffer{}

	err = t.Execute(buf, map[string]string{
		"From": fmt.Sprintf("<@%s>", from),
		"To":   fmt.Sprintf("<@%s>", to),
	})
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}

func parseUserGroupID(s string) (string, bool) {
	r, err := regexp.Compile("<!subteam\\^(.*)\\|@.*>")
	if err != nil {
		panic(err.Error())
	}

	submatches := r.FindStringSubmatch(s)
	if submatches == nil || len(submatches) < r.NumSubexp() {
		return "", false
	}

	return submatches[1], true
}

func parseUserID(s string) (string, bool) {
	r, err := regexp.Compile("<@(.*)>")
	if err != nil {
		panic(err.Error())
	}

	submatches := r.FindStringSubmatch(s)
	if submatches == nil || len(submatches) < r.NumSubexp() {
		return "", false
	}

	return submatches[1], true
}

var errMessage = "Sorry, something has gone wrong!"

func hadError(err error, w http.ResponseWriter, message string) bool {
	if err == nil {
		return false
	}
	handleError(err, w, message)
	return true
}

func handleError(err error, w http.ResponseWriter, message string) {
	log.Print(err.Error())
	_, err = w.Write([]byte(message))
	if err != nil {
		log.Print(err.Error())
	}
}
