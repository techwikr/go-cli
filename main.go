package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"unicode"
	"unicode/utf8"
)

type Event struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
	Payload json.RawMessage `json:"payload"`
}

type PushPayload struct {
	Commits []struct {
		Message string `json:"message"`
	} `json:"commits"`
}

type CreatePayload struct {
	RefType string `json:"ref_type"`
}

type DeletePayload struct {
	RefType string `json:"ref_type"`
}

type IssuesPayload struct {
	Action string `json:"action"`
}

type PullRequestPayload struct {
	Action string `json:"action"`
}

func formatEvent(event Event) string {
	timeStr := event.CreatedAt.Format("2006-01-02 15:04")
	repo := event.Repo.Name

	switch event.Type {
	case "PushEvent":
		var payload PushPayload
		_ = json.Unmarshal(event.Payload, &payload)
		count := len(payload.Commits)
		commitWord := "commit"
		if count != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("🔨 Pushed %d %s to %s (%s)", count, commitWord, repo, timeStr)

	case "CreateEvent":
		var payload CreatePayload
		_ = json.Unmarshal(event.Payload, &payload)
		refType := payload.RefType
		if refType == "" {
			refType = "repository"
		}
		return fmt.Sprintf("✨ Created %s in %s (%s)", refType, repo, timeStr)

	case "DeleteEvent":
		var payload DeletePayload
		_ = json.Unmarshal(event.Payload, &payload)
		refType := payload.RefType
		if refType == "" {
			refType = "branch"
		}
		return fmt.Sprintf("🗑️  Deleted %s in %s (%s)", refType, repo, timeStr)

	case "IssuesEvent":
		var payload IssuesPayload
		_ = json.Unmarshal(event.Payload, &payload)
		action := payload.Action
		if action == "" {
			action = "updated"
		}
		return fmt.Sprintf("📝 %s an issue in %s (%s)", capitalize(action), repo, timeStr)

	case "IssueCommentEvent":
		return fmt.Sprintf("💬 Commented on an issue in %s (%s)", repo, timeStr)

	case "WatchEvent":
		return fmt.Sprintf("⭐ Starred %s (%s)", repo, timeStr)

	case "ForkEvent":
		return fmt.Sprintf("🍴 Forked %s (%s)", repo, timeStr)

	case "PullRequestEvent":
		var payload PullRequestPayload
		_ = json.Unmarshal(event.Payload, &payload)
		action := payload.Action
		if action == "" {
			action = "updated"
		}
		return fmt.Sprintf(" %s a pull request in %s (%s)", capitalize(action), repo, timeStr)

	case "PullRequestReviewEvent":
		return fmt.Sprintf(" Reviewed a pull request in %s (%s)", repo, timeStr)

	case "PullRequestReviewCommentEvent":
		return fmt.Sprintf(" Commented on a pull request in %s (%s)", repo, timeStr)

	case "ReleaseEvent":
		return fmt.Sprintf(" Published a release in %s (%s)", repo, timeStr)

	case "MemberEvent":
		return fmt.Sprintf(" Added a collaborator to %s (%s)", repo, timeStr)

	default:
		eventName := event.Type
		if len(eventName) > 5 && eventName[len(eventName)-5:] == "Event" {
			eventName = eventName[:len(eventName)-5]
		}
		return fmt.Sprintf("📌 %s in %s (%s)", eventName, repo, timeStr)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func fetchUserActivity(username string) error {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	fmt.Printf("Fetching activity for GitHub user: %s\n\n", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "github-activity-fetcher")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("user '%s' not found", username)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(events) == 0 {
		fmt.Println("No recent activity found for this user.")
		return nil
	}

	fmt.Printf("Recent activity for %s:\n", username)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, event := range events {
		fmt.Println(formatEvent(event))
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-activity <username>")
		fmt.Println("\nExample:")
		fmt.Println("  github-activity torvalds")
		os.Exit(1)
	}

	username := os.Args[1]

	if err := fetchUserActivity(username); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
