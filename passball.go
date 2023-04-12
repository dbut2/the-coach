package football

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/slack-go/slack"
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

	userGroupID, is := parseUserGroupID(split[0])
	if !is {
		handleError(fmt.Errorf("invalid usergroup: %s", split[0]), w, "Please supply a valid user group")
		return
	}

	if len(split) < 2 {
		handleError(errors.New("no receiver supplied"), w, "Please supply at least one user or user group receiver")
		return
	}

	sc := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	var candidates []string

	for _, receiver := range split[1:] {
		parsed := false

		if receiverUserGroupID, is := parseUserGroupID(receiver); is {
			userIDs, err := sc.GetUserGroupMembers(receiverUserGroupID)
			if hadError(err, w, errMessage) {
				return
			}

			parsed = true
			candidates = union(candidates, userIDs)
		}

		if receiverUserID, is := parseUserID(receiver); is {
			parsed = true
			candidates = union(candidates, []string{receiverUserID})
		}

		if receiverChannelID, is := parseChannelID(receiver); is {
			userIDs, _, err := sc.GetUsersInConversation(&slack.GetUsersInConversationParameters{
				ChannelID: receiverChannelID,
			})
			if hadError(err, w, errMessage) {
				return
			}

			parsed = true
			candidates = union(candidates, userIDs)
		}

		if !parsed {
			handleError(fmt.Errorf("invalid receiver: %s", receiver), w, "Please supply valid user or user group receivers")
			return
		}
	}

	if len(candidates) == 0 {
		handleError(errors.New("no candidates found"), w, "Sorry, no potential recruits were found")
		return
	}

	if len(candidates) > 1 {
		candidates = remove(candidates, s.UserID)
	}

	newUserID := candidates[rand.Intn(len(candidates))]

	_, _, err := sc.PostMessage(s.ChannelID, slack.MsgOptionText(randomPhrase(s.UserID, newUserID, userGroupID), false))
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

func randomPhrase(from, to, group string) string {
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
		"From":  fmt.Sprintf("<@%s>", from),
		"To":    fmt.Sprintf("<@%s>", to),
		"Group": fmt.Sprintf("<!subteam^%s>", group),
	})
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}

func parseUserGroupID(s string) (string, bool) {
	r, err := regexp.Compile("<!subteam\\^(S[A-Z0-9]*)(\\|.*)?>")
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
	r, err := regexp.Compile("<@(U[A-Z0-9]*)(\\|.*)?>")
	if err != nil {
		panic(err.Error())
	}

	submatches := r.FindStringSubmatch(s)
	if submatches == nil || len(submatches) < r.NumSubexp() {
		return "", false
	}

	return submatches[1], true
}

func parseChannelID(s string) (string, bool) {
	r, err := regexp.Compile("<#(C[A-Z0-9]*)(\\|.*)?>")
	if err != nil {
		panic(err.Error())
	}

	submatches := r.FindStringSubmatch(s)
	if submatches == nil || len(submatches) < r.NumSubexp() {
		return "", false
	}

	return submatches[1], true
}

func union(sss ...[]string) []string {
	m := make(map[string]bool)
	for _, ss := range sss {
		for _, s := range ss {
			m[s] = true
		}
	}
	var l []string
	for s := range m {
		l = append(l, s)
	}
	return l
}

func remove(s []string, item string) []string {
	var result []string
	for _, str := range s {
		if str == item {
			continue
		}
		result = append(result, str)
	}
	return result
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
