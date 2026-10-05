package sanitize

import "github.com/microcosm-cc/bluemonday"

// bioPolicy strips all HTML from user-supplied free-text fields. Use
// StrictPolicy (no tags allowed) unless you deliberately support rich
// text, in which case use UGCPolicy() and store the sanitized output,
// never the raw input.
var plainTextPolicy = bluemonday.StrictPolicy()

// SanitizeText strips any HTML/script content from a user-supplied string.
// Call this in the service layer before persisting fields like bio,
// display name, or job descriptions that came from user input.
func PlainText(input string) string {
	return plainTextPolicy.Sanitize(input)
}
