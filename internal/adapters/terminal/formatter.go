package terminal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type presenter struct {
	w io.Writer
}

func NewPresenter(w io.Writer) core.Presenter {
	return &presenter{w: w}
}

func (p *presenter) ShowResponse(resp core.Response, opts core.PresentOpts) error {
	if opts.Raw {
		_, err := fmt.Fprintf(p.w, "%s\n", resp.Body)
		return err
	}

	fmt.Fprintf(p.w, "Status: %d %s\n", resp.StatusCode, resp.StatusText)
	fmt.Fprintf(p.w, "Time:   %dms\n", resp.Duration)

	if opts.Verbose {
		fmt.Fprintf(p.w, "\nResponse Headers:\n")
		for _, k := range sortedKeys(resp.Headers) {
			fmt.Fprintf(p.w, "  %s: %s\n", k, resp.Headers[k])
		}
	}

	fmt.Fprintf(p.w, "\n")
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(resp.Body), "", "  "); err != nil {
		fmt.Fprintf(p.w, "%s\n", highlightJSON(resp.Body))
		return nil
	}
	fmt.Fprintf(p.w, "%s\n", highlightJSON(pretty.String()))
	return nil
}

func (p *presenter) ShowRequestDetails(req core.Request) error {
	fmt.Fprintf(p.w, "Method: %s\n", req.Method)
	fmt.Fprintf(p.w, "URL:    %s\n", req.URL)

	if len(req.Headers) > 0 {
		fmt.Fprintf(p.w, "\nHeaders:\n")
		for _, k := range sortedKeys(req.Headers) {
			fmt.Fprintf(p.w, "  %s: %s\n", k, maskSensitive(k, req.Headers[k]))
		}
	}

	return nil
}

func (p *presenter) ShowCollections(collections []core.Collection) error {
	fmt.Fprintf(p.w, "Collections:\n")
	for _, c := range collections {
		fmt.Fprintf(p.w, "  %s (%s)\n", c.Name, c.Path)
	}
	return nil
}

func (p *presenter) ShowRequests(requests []core.Request) error {
	fmt.Fprintf(p.w, "Requests:\n")
	for _, r := range requests {
		fmt.Fprintf(p.w, "  %-7s%s\n", r.Method, r.Name)
	}
	return nil
}

func (p *presenter) ShowError(msg string) error {
	fmt.Fprintf(p.w, "Error: %s", msg)
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func maskSensitive(key, value string) string {
	lower := strings.ToLower(key)
	sensitive := []string{"auth", "token", "key", "secret", "password"}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return "***"
		}
	}
	return value
}

type StdoutWriter struct{}

func NewStdoutWriter() *StdoutWriter {
	return &StdoutWriter{}
}

func (s *StdoutWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}
