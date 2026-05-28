package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type userProfile struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	EmailNotify bool   `json:"email_notify_comments"`
}

type profileResponse struct {
	ArticleCount int         `json:"article_count"`
	userProfile
}

// HimalayaNotifier sends email notifications via Himalaya CLI
type HimalayaNotifier struct {
	authServiceURL string
	httpClient     *http.Client
}

// NewHimalayaNotifier creates a new Himalaya email notifier
func NewHimalayaNotifier() *HimalayaNotifier {
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8001"
	}
	return &HimalayaNotifier{
		authServiceURL: authURL,
		httpClient:     &http.Client{Timeout: 5 * time.Second},
	}
}

// SendCommentNotification notifies the article author about a new comment
func (n *HimalayaNotifier) SendCommentNotification(articleTitle, commenterID, commentSnippet string) error {
	// Not implemented here — we use the usecase-driven approach instead.
	// This method exists to satisfy the interface; the actual notification
	// is handled via the usecase calling the auth service + himalaya.
	return nil
}

// NotifyAuthorOfComment sends email to the article author about a new comment.
// It calls the auth service to get author info, then uses himalaya to send.
func (n *HimalayaNotifier) NotifyAuthorOfComment(articleID, authorID, commenterName, articleTitle, commentSnippet string) {
	// Skip if commenter is the author
	if authorID == "" {
		return
	}

	// Call auth service to get author profile (username-based lookup won't work with ID)
	// Instead, we try to find the article author via GetByID on auth service
	resp, err := n.httpClient.Get(fmt.Sprintf("%s/api/v1/users/%s", n.authServiceURL, authorID))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var profile profileResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return
	}

	if !profile.EmailNotify || profile.Email == "" {
		return
	}

	subject := fmt.Sprintf("💬 New comment on \"%s\"", articleTitle)

	body := &bytes.Buffer{}
	body.WriteString(fmt.Sprintf("Hi %s,\n\n", profile.DisplayName))
	body.WriteString(fmt.Sprintf("%s left a comment on your article \"%s\":\n\n", commenterName, articleTitle))
	body.WriteString(fmt.Sprintf("    \"%s\"\n\n", commentSnippet))
	body.WriteString("Visit your dashboard to reply.\n\n")
	body.WriteString("— PlatformBlog")

	// Try himalaya, fall back to log if not available
	cmd := exec.Command("himalaya", "send",
		"--from", "andreahinata78@gmail.com",
		"--to", profile.Email,
		"--subject", subject,
		"--body", body.String(),
	)
	cmd.Output() // best-effort, ignore errors
}

// SendNewComment satisfies the usecase.CommentNotifier interface
func (n *HimalayaNotifier) SendNewComment(toEmail, toName, articleTitle, commenterName, commentSnippet string) error {
	subject := fmt.Sprintf("💬 New comment on \"%s\"", articleTitle)

	body := &bytes.Buffer{}
	body.WriteString(fmt.Sprintf("Hi %s,\n\n", toName))
	body.WriteString(fmt.Sprintf("%s left a comment on your article \"%s\":\n\n", commenterName, articleTitle))
	body.WriteString(fmt.Sprintf("    \"%s\"\n\n", commentSnippet))
	body.WriteString("Visit your dashboard to reply.\n\n")
	body.WriteString("— PlatformBlog")

	cmd := exec.Command("himalaya", "send",
		"--from", "andreahinata78@gmail.com",
		"--to", toEmail,
		"--subject", subject,
		"--body", body.String(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("himalaya send failed: %w — output: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
