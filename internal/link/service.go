package link

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

var tracer = otel.Tracer("github.com/urlspace/api/internal/link")

var errBlockedAddress = errors.New("blocked address")

const (
	metadataMaxInFlight  = 20
	metadataMaxBytes     = 512 << 10
	metadataMaxHeaders   = 64 << 10
	metadataTimeout      = 5 * time.Second
	metadataMaxRedirects = 3
	metadataUserAgent    = "url.space"
	metadataMaxTitle     = 255
	metadataMinTitle     = 3
	metadataMaxDesc      = 512
)

type oembedProvider struct {
	host        string
	endpoint    string
	authorLabel string
}

// Hosts that serve our servers a generic page, so every link would get the
// same misleading title. Their oEmbed endpoint returns the real one instead.
// Subdomains are included. The endpoint is fixed, so the user's URL only
// reaches the provider as a query parameter, never as the address we dial.
var oembedProviders = []oembedProvider{
	{host: "youtube.com", endpoint: "https://www.youtube.com/oembed", authorLabel: "YouTube video by "},
	{host: "youtu.be", endpoint: "https://www.youtube.com/oembed", authorLabel: "YouTube video by "},
	{host: "reddit.com", endpoint: "https://www.reddit.com/oembed", authorLabel: "Reddit post by "},
}

// Ranges the metadata fetcher must never connect to, on top of the checks in
// isBlockedAddr. Without them, a URL like http://127.0.0.1.nip.io/ (a public
// name that resolves to loopback) would let anyone reach services on our own
// network through the API.
var blockedPrefixes = []netip.Prefix{
	// "This network": 0.0.0.0/8, e.g. 0.1.2.3. Some systems route it to
	// the local machine.
	netip.MustParsePrefix("0.0.0.0/8"),
	// Carrier-grade NAT, used inside cloud networks, e.g. 100.64.1.5.
	netip.MustParsePrefix("100.64.0.0/10"),
	// IETF protocol assignments, e.g. 192.0.0.8.
	netip.MustParsePrefix("192.0.0.0/24"),
	// Benchmarking networks, e.g. 198.18.0.1.
	netip.MustParsePrefix("198.18.0.0/15"),
	// Reserved for future use, e.g. 240.0.0.1, including the broadcast
	// address 255.255.255.255.
	netip.MustParsePrefix("240.0.0.0/4"),
	// NAT64: an IPv6 address that wraps an IPv4 one, e.g. 64:ff9b::7f00:1 is
	// 127.0.0.1 on networks that translate it.
	netip.MustParsePrefix("64:ff9b::/96"),
	// Local-use NAT64, the same trick with a network's own prefix, e.g.
	// 64:ff9b:1::7f00:1.
	netip.MustParsePrefix("64:ff9b:1::/48"),
	// Old IPv4-compatible IPv6, e.g. ::7f00:1 is 127.0.0.1 on some systems.
	netip.MustParsePrefix("::/96"),
	// 6to4 tunnels, which carry an IPv4 address inside, e.g. 2002:7f00:1:: is
	// 127.0.0.1.
	netip.MustParsePrefix("2002::/16"),
	// Teredo tunnels, which also carry an IPv4 address, e.g. 2001:0:7f00:1::.
	netip.MustParsePrefix("2001::/32"),
	// Discard-only, traffic is dropped, e.g. 100::1.
	netip.MustParsePrefix("100::/64"),
	// Documentation examples, never real hosts, e.g. 2001:db8::1.
	netip.MustParsePrefix("2001:db8::/32"),
}

type Metadata struct {
	Title       string
	Description string
}

type Service struct {
	client *http.Client
	slots  chan struct{}
}

func NewService() *Service {
	dialer := &net.Dialer{
		Control: checkAddress,
	}
	transport := &http.Transport{
		// Never pick up HTTP_PROXY and friends from the environment.
		Proxy:                  nil,
		DialContext:            dialer.DialContext,
		ForceAttemptHTTP2:      true,
		MaxResponseHeaderBytes: metadataMaxHeaders,
		// Fetches almost never hit the same host twice, so there's nothing to reuse.
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= metadataMaxRedirects {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("redirect to a non-http scheme")
			}
			trace.SpanFromContext(req.Context()).SetAttributes(attribute.Int("metadata.redirects", len(via)))
			return nil
		},
	}
	return &Service{
		client: client,
		slots:  make(chan struct{}, metadataMaxInFlight),
	}
}

// FetchMetadata returns the title and description of the page at rawURL, which
// must already have passed ValidateURL. Any failure, including too many fetches
// in flight, returns empty Metadata, so callers can't tell an internal address
// from a site that simply has no tags.
func (s *Service) FetchMetadata(ctx context.Context, userID uuid.UUID, rawURL string, acceptLanguage string) Metadata {
	ctx, span := tracer.Start(ctx, "link.FetchMetadata")
	defer span.End()

	outcome := "error"
	defer func() {
		span.SetAttributes(attribute.String("metadata.outcome", outcome))
	}()

	u, err := url.Parse(rawURL)
	if err != nil {
		return Metadata{}
	}

	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	default:
		slog.WarnContext(ctx, "link metadata concurrency cap reached",
			slog.String("domain", "link"),
			slog.Int("limit", metadataMaxInFlight),
			slog.String("user_id", userID.String()),
		)
		outcome = "busy"
		return Metadata{}
	}

	ctx, cancel := context.WithTimeout(ctx, metadataTimeout)
	defer cancel()

	if provider, ok := findOembedProvider(u.Hostname()); ok {
		span.SetAttributes(attribute.String("metadata.source", "oembed"))
		var metadata Metadata
		metadata, outcome = s.fetchOembed(ctx, provider, u.String(), acceptLanguage)
		return metadata
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Metadata{}
	}
	req.Header.Set("User-Agent", metadataUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}

	res, err := s.client.Do(req)
	if err != nil {
		var netErr net.Error
		switch {
		case errors.Is(err, errBlockedAddress):
			outcome = "blocked_ip"
		case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
			outcome = "timeout"
		}
		return Metadata{}
	}
	defer res.Body.Close()

	span.SetAttributes(attribute.Int("http.response.status_code", res.StatusCode))
	if res.StatusCode != http.StatusOK {
		outcome = "bad_status"
		return Metadata{}
	}

	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || (mediaType != "text/html" && mediaType != "application/xhtml+xml") {
		outcome = "not_html"
		return Metadata{}
	}

	// Converts pages in other encodings (e.g. windows-1250) to UTF-8, using
	// the Content-Type header or a <meta charset> near the top of the page.
	body, err := charset.NewReader(io.LimitReader(res.Body, metadataMaxBytes), contentType)
	if err != nil {
		return Metadata{}
	}

	metadata := extractMetadata(body)
	if metadata.Title == "" && metadata.Description == "" {
		outcome = "empty"
	} else {
		outcome = "ok"
	}
	return metadata
}

// checkAddress runs for every connection the fetcher opens, including each
// redirect, after DNS has resolved the host. Checking the real IP here, not
// the hostname in the URL, is what stops DNS tricks and redirects to internal
// addresses.
func checkAddress(network string, address string, _ syscall.RawConn) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return errBlockedAddress
	}
	// Real pages live on the standard ports. Anything else, e.g.
	// example.com:8080, could be probing an internal service.
	if port != "80" && port != "443" {
		return errBlockedAddress
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return errBlockedAddress
	}
	if isBlockedAddr(addr) {
		return errBlockedAddress
	}
	return nil
}

func isBlockedAddr(addr netip.Addr) bool {
	// An IPv4 address written as IPv6, e.g. ::ffff:127.0.0.1, is checked as
	// the IPv4 address it wraps.
	addr = addr.Unmap()

	switch {
	// Loopback, the machine itself: 127.0.0.1, ::1.
	case addr.IsLoopback():
		return true
	// Private networks: 10.0.0.5, 172.16.0.1, 192.168.1.1, and IPv6 fc00::/7
	// such as fd12::1, which is what Railway's private network uses.
	case addr.IsPrivate():
		return true
	// Link-local: 169.254.169.254 (the cloud metadata endpoint), fe80::1.
	case addr.IsLinkLocalUnicast(), addr.IsLinkLocalMulticast():
		return true
	// Unspecified, "any address": 0.0.0.0, ::.
	case addr.IsUnspecified():
		return true
	// Multicast, not a single host: 224.0.0.1, ff02::1.
	case addr.IsMulticast(), addr.IsInterfaceLocalMulticast():
		return true
	}

	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func findOembedProvider(host string) (oembedProvider, bool) {
	host = strings.ToLower(host)
	for _, provider := range oembedProviders {
		if host == provider.host || strings.HasSuffix(host, "."+provider.host) {
			return provider, true
		}
	}
	return oembedProvider{}, false
}

// fetchOembed asks the provider for the title and author of rawURL. It uses the
// same client as page fetches, so the address checks and redirect limits still
// apply. It returns the metadata and the outcome to record on the span.
func (s *Service) fetchOembed(ctx context.Context, provider oembedProvider, rawURL string, acceptLanguage string) (Metadata, string) {
	endpoint, err := url.Parse(provider.endpoint)
	if err != nil {
		return Metadata{}, "error"
	}
	endpoint.RawQuery = url.Values{"format": {"json"}, "url": {rawURL}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Metadata{}, "error"
	}
	req.Header.Set("User-Agent", metadataUserAgent)
	req.Header.Set("Accept", "application/json")
	if acceptLanguage != "" {
		req.Header.Set("Accept-Language", acceptLanguage)
	}

	res, err := s.client.Do(req)
	if err != nil {
		var netErr net.Error
		switch {
		case errors.Is(err, errBlockedAddress):
			return Metadata{}, "blocked_ip"
		case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
			return Metadata{}, "timeout"
		}
		return Metadata{}, "error"
	}
	defer res.Body.Close()

	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("http.response.status_code", res.StatusCode))
	if res.StatusCode != http.StatusOK {
		return Metadata{}, "bad_status"
	}

	var body struct {
		Title      string `json:"title"`
		AuthorName string `json:"author_name"`
	}
	if err := json.UnmarshalRead(io.LimitReader(res.Body, metadataMaxBytes), &body); err != nil {
		return Metadata{}, "error"
	}

	title := clean(body.Title, metadataMaxTitle)
	if utf8.RuneCountInString(title) < metadataMinTitle {
		title = ""
	}
	var description string
	if author := clean(body.AuthorName, metadataMaxDesc); author != "" {
		description = clean(provider.authorLabel+author, metadataMaxDesc)
	}
	if title == "" && description == "" {
		return Metadata{}, "empty"
	}
	return Metadata{Title: title, Description: description}, "ok"
}

// extractMetadata reads tags from the <head> only, stopping at </head> or
// when the reader runs out (it is capped at metadataMaxBytes). og: tags win
// over <title> and <meta name="description">. The tokenizer decodes entities
// like &rsquo; in both text and attribute values.
func extractMetadata(r io.Reader) Metadata {
	var ogTitle, title, ogDescription, description string

	z := html.NewTokenizer(r)
loop:
	for {
		switch z.Next() {
		case html.ErrorToken:
			break loop
		case html.EndTagToken:
			if name, _ := z.TagName(); string(name) == "head" {
				break loop
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "title":
				if title == "" && z.Next() == html.TextToken {
					title = string(z.Text())
				}
			case "meta":
				if !hasAttr {
					continue
				}
				var property, metaName, content string
				for {
					key, val, more := z.TagAttr()
					switch string(key) {
					case "property":
						property = string(val)
					case "name":
						metaName = string(val)
					case "content":
						content = string(val)
					}
					if !more {
						break
					}
				}
				switch {
				case property == "og:title" && ogTitle == "":
					ogTitle = content
				case property == "og:description" && ogDescription == "":
					ogDescription = content
				case strings.EqualFold(metaName, "description") && description == "":
					description = content
				}
			}
		}
	}

	cleanTitle := clean(firstNonEmpty(ogTitle, title), metadataMaxTitle)
	if utf8.RuneCountInString(cleanTitle) < metadataMinTitle {
		cleanTitle = ""
	}
	return Metadata{
		Title:       cleanTitle,
		Description: clean(firstNonEmpty(ogDescription, description), metadataMaxDesc),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// clean turns control characters into spaces, collapses whitespace and cuts
// the text to max characters (not bytes, so a letter like ą is never split).
func clean(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if runes := []rune(s); len(runes) > max {
		s = strings.TrimSpace(string(runes[:max]))
	}
	return s
}
