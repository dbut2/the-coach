package football

import (
	"bytes"
	_ "embed"
	"encoding/json"
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

	"github.com/slack-go/slack"
)

var (
	slackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET")
	slackBotToken      = os.Getenv("SLACK_BOT_TOKEN")
)

func PassBall(w http.ResponseWriter, r *http.Request) {
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

	err = handleRequestSlash(s)
	if hadError(err, w, fmt.Sprintf("%s: %s", errMessage, err.Error())) {
		return
	}
}

func handleRequestSlash(s slack.SlashCommand) error {
	sc := slack.New(slackBotToken)

	if s.Text != "" {
		msg, err := prepareDescriptionMessage(s.Text)
		if err != nil {
			return err
		}

		_, err = sc.PostEphemeral(s.ChannelID, s.UserID, slack.MsgOptionText(msg, true))
		if err != nil {
			return err
		}

		return nil
	}

	return passBall(sc, s)
}

type config struct {
	From string   `json:"from"`
	To   []string `json:"to"`
}

func prepareDescriptionMessage(args string) (string, error) {
	split := strings.Split(args, " ")
	if len(split) < 2 {
		return "Please enter a user group to pass from and one to pass to", nil
	}

	c := &config{
		From: split[0],
		To:   split[1:],
	}

	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Add the following line to the channel description:\ncoach-peter: %s", string(b)), nil
}

func passBall(sc *slack.Client, s slack.SlashCommand) error {
	info, err := sc.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: s.ChannelID,
	})
	if err != nil {
		return err
	}

	c, err := parseDescription(info.Purpose)
	if err != nil {
		return err
	}

	userGroupID, is := parseUserGroupID(c.From)
	if !is {
		return errors.New("invalid user group supplied")
	}

	candidates, err := getCandidates(sc, c.To)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return errors.New("no candidates found")
	}

	if len(candidates) > 1 {
		candidates = remove(candidates, s.UserID)
	}

	newUserID := candidates[rand.Intn(len(candidates))]

	_, _, err = sc.PostMessage(s.ChannelID, slack.MsgOptionText(randomPhrase(s.UserID, newUserID, userGroupID), false))
	if err != nil {
		return err
	}

	_, err = sc.UpdateUserGroupMembers(userGroupID, newUserID)
	if err != nil {
		return err
	}

	return nil
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

func getCandidates(sc *slack.Client, receivers []string) ([]string, error) {
	var candidates []string

	for _, receiver := range receivers {
		parsed := false

		if receiverUserGroupID, is := parseUserGroupID(receiver); is {
			userIDs, err := sc.GetUserGroupMembers(receiverUserGroupID)
			if err != nil {
				return nil, err
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
			if err != nil {
				return nil, err
			}

			parsed = true
			candidates = union(candidates, userIDs)
		}

		if !parsed {
			return nil, fmt.Errorf("invalid receiver: %s", receiver)
		}
	}
	return candidates, nil
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

func parseDescription(description slack.Purpose) (*config, error) {
	for _, line := range strings.Split(description.Value, "\n") {
		split := strings.SplitN(line, ":", 2)
		if len(split) < 2 {
			continue
		}

		if strings.TrimSpace(split[0]) != "coach-peter" {
			continue
		}

		b := []byte(strings.TrimSpace(split[1]))
		c := new(config)
		err := json.Unmarshal(b, c)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return nil, errors.New("no config found in description")
}

var errMessage = "Sorry, something has gone wrong"

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
