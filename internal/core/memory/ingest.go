package memory

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

// ingest.go — the ingest flow from 07-memory.md §1:
// probe -> licence-detect -> distil (injected seam) -> validate -> atomic write
// -> discard/keep-original. Every pre-dispatch failure raises before the single
// WritePages call — no orphan registry row, no partial page.

const ingestedBy = "abcd memory ingest"

const (
	maxFetchBytes       = 10 * 1024 * 1024
	fetchTimeoutSeconds = 30
	maxRedirects        = 5
)

var textContentTypes = map[string]bool{
	"application/json": true, "application/xml": true, "application/xhtml+xml": true,
}

var extRe = regexp.MustCompile(`^\.[a-z0-9]{1,10}$`)

// FetchedSource is the raw result of fetching a URL (the injectable Fetcher
// contract). Content-type / size / decode checks are applied uniformly by the
// ingest path after the fetcher returns.
type FetchedSource struct {
	FinalURL string
	Headers  map[string]string
	Body     []byte
}

// Fetcher fetches a URL; a nil Fetcher uses the bounded default fetch.
type Fetcher func(url string) (FetchedSource, error)

// PDFExtractor extracts text from PDF bytes; a nil extractor rejects with a
// clear error (never silently pulls in a parser dependency).
type PDFExtractor func(data []byte) (string, error)

// Distiller is the host-delegated seam: (normalisedText, sourceBlock) -> raw
// page maps. A map omitting "source" gets sourceBlock injected. The core
// validates every page.
type Distiller func(normalisedText string, sourceBlock map[string]any) ([]map[string]any, error)

// IngestRequest is the input to Ingest.
type IngestRequest struct {
	RepoRoot     string
	Source       string
	Distiller    Distiller
	KeepOriginal bool
	Fetcher      Fetcher
	PDFExtractor PDFExtractor
	Now          time.Time
}

// IngestResult is the structured result of one Ingest call.
type IngestResult struct {
	Status           string         `json:"status"`
	ContentHash      string         `json:"content_hash"`
	Licence          string         `json:"licence"`
	SourceTokenCount int            `json:"source_token_count"`
	Pages            []string       `json:"pages"`
	Citation         map[string]any `json:"citation"`
	KeptOriginal     string         `json:"kept_original"`
	// KeepOriginalError records a --keep-original copy failure that occurred
	// AFTER the pages and registry were durably written. The ingest itself
	// succeeded; only the best-effort original copy did not. Empty when
	// --keep-original was not requested or the copy succeeded.
	KeepOriginalError string       `json:"keep_original_error,omitempty"`
	Linked            [][2]string  `json:"linked"`
	Contradictions    [][2]string  `json:"contradictions"`
	WriteReport       *WriteReport `json:"write_report"`
}

type sourceMaterial struct {
	origin      string
	text        string
	rawBytes    []byte
	headers     map[string]string
	ext         string
	sourceClass string
	title       string
}

// Ingest runs the full ingest flow for one source (local path or http(s) URL).
func Ingest(req IngestRequest) (IngestResult, error) {
	if req.Distiller == nil {
		return IngestResult{}, newIngestError("ingest requires a Distiller")
	}
	root := req.RepoRoot
	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	ingestedAt := now.Format("2006-01-02")

	material, err := acquireSource(req.Source, req.Fetcher, req.PDFExtractor)
	if err != nil {
		return IngestResult{}, err
	}
	normalized := NormaliseSourceText(material.text)
	if strings.TrimSpace(normalized) == "" {
		return IngestResult{}, newIngestError("source has no text content: %s", material.origin)
	}
	contentHash := SourceContentHash(material.text)
	tokenCount := CountSourceTokens(normalized)

	registry, err := LoadRegistry(SourcesIndexPath(root))
	if err != nil {
		return IngestResult{}, err
	}
	entry, _ := registry[contentHash].(map[string]any)
	var memoryConsumer map[string]any
	if entry != nil {
		if consumers, ok := entry["consumers"].(map[string]any); ok {
			memoryConsumer, _ = consumers["memory"].(map[string]any)
		}
	}
	mem := Dir(root)

	// ---- Registry-hit fast path (validate BEFORE mutate) -------------------
	var validRecorded []string
	var recorded []string
	if memoryConsumer != nil {
		recorded = anyToStrings(memoryConsumer["pages"])
		allValid := len(recorded) > 0
		for _, pageName := range recorded {
			hashes, present := pageHashSet(mem, pageName)
			if present && contains(hashes, contentHash) {
				validRecorded = append(validRecorded, pageName)
			} else {
				allValid = false
			}
		}
		if allValid {
			sourceClass := material.sourceClass
			if c, ok := memoryConsumer["class"].(string); ok && c != "" {
				sourceClass = c
			}
			citation, _ := memoryConsumer["citation"].(map[string]any)
			if citation == nil {
				citation = map[string]any{}
			}
			origin := material.origin
			if o, ok := entry["origin"].(string); ok && o != "" {
				origin = o
			}
			licence := "unknown"
			if l, ok := entry["licence"].(string); ok && l != "" {
				licence = l
			}
			event := IngestEvent{
				ContentHash: contentHash, Consumer: "memory", SourceClass: sourceClass,
				Citation: citation, Origin: origin, Licence: licence, IngestedAt: ingestedAt,
				Pages: recorded, SourceTokenCount: tokenCount, TokenCountVersion: TokenCountVersion,
			}
			var newRegistry map[string]any
			report, err := WritePages(root, nil, func(current map[string]any) (map[string]any, error) {
				merged, err := MergeIngest(current, event)
				if err != nil {
					return nil, err
				}
				newRegistry = merged
				return merged, nil
			}, now)
			if err != nil {
				return IngestResult{}, err
			}
			// Best-effort keep-original: a failure after the durable write is
			// recorded, never reported as total failure (iss-30).
			kept, keepErr := "", ""
			if req.KeepOriginal {
				if k, serr := storeOriginal(root, material, contentHash); serr != nil {
					keepErr = keepOriginalErrorMessage(serr)
				} else {
					kept = k
				}
			}
			cachedEntry, _ := newRegistry[contentHash].(map[string]any)
			cachedConsumers, _ := cachedEntry["consumers"].(map[string]any)
			cached, _ := cachedConsumers["memory"].(map[string]any)
			cachedCitation, _ := cached["citation"].(map[string]any)
			resultLicence := "unknown"
			if l, ok := cachedEntry["licence"].(string); ok && l != "" {
				resultLicence = l
			}
			return IngestResult{
				Status: "registry_only", ContentHash: contentHash, Licence: resultLicence,
				SourceTokenCount: tokenCount, Pages: recorded, Citation: deepCopyMap(cachedCitation),
				KeptOriginal: kept, KeepOriginalError: keepErr, WriteReport: &report,
			}, nil
		}
	}

	repairing := memoryConsumer != nil

	// ---- Licence detect (sourceRoot="": SPDX header + HTTP License:) --------
	detection := DetectLicence(material.text, "", material.headers)
	licence := detection.Licence

	citation := BuildCitation("knowledge", material.origin, "unknown", material.title, now.Year(), ingestedAt, ingestedBy)
	sourceBlock, err := buildSingleSource(material.sourceClass, citation, licence, contentHash, ingestedAt)
	if err != nil {
		return IngestResult{}, err
	}

	// ---- Distil + validate BEFORE any write --------------------------------
	rawPages, err := req.Distiller(normalized, sourceBlock)
	if err != nil {
		return IngestResult{}, err
	}
	var distilled []DistilledPage
	for _, raw := range rawPages {
		if _, ok := raw["source"]; !ok {
			merged := map[string]any{}
			for k, v := range raw {
				merged[k] = v
			}
			merged["source"] = sourceBlock
			raw = merged
		}
		page, err := ValidateDistilledPage(raw)
		if err != nil {
			return IngestResult{}, err
		}
		distilled = append(distilled, page)
	}
	if len(distilled) == 0 {
		return IngestResult{}, newIngestError("distillation produced 0 pages for %s; nothing written", material.origin)
	}
	for _, page := range distilled {
		if !contains(SourceHashes(page.Source), contentHash) {
			return IngestResult{}, newIngestError("distilled page %s does not cite the ingested source hash %s; refusing to write an unattributable page", page.Filename(), contentHash)
		}
	}

	// ---- Existing pages + repair safety ------------------------------------
	existing := existingPageFrontmatter(mem)
	if repairing {
		for _, pageName := range recorded {
			if contains(validRecorded, pageName) {
				continue
			}
			hashes, present := pageHashSet(mem, pageName)
			if !present {
				continue // missing — re-distil writes fresh
			}
			if len(hashes) == 0 {
				delete(existing, pageName)
				continue
			}
			for _, h := range hashes {
				if _, ok := registry[h]; ok {
					return IngestResult{}, newIngestError("repair collision: recorded page %s now cites a different registry entry; operator resolution required, nothing overwritten", pageName)
				}
			}
		}
	}

	plan, err := ResolveDistilledPages(existing, distilled)
	if err != nil {
		return IngestResult{}, err
	}

	// ---- Build the COMPLETE new registry mapping ---------------------------
	ourPages := append([]string(nil), plan.RegistryPages[contentHash]...)
	for _, pageName := range validRecorded {
		if !contains(ourPages, pageName) {
			ourPages = append(ourPages, pageName)
		}
	}
	// Back-link invariant: the entry lists EXACTLY the live page set.
	dedupPages := dedupStrings(ourPages)
	event := IngestEvent{
		ContentHash: contentHash, Consumer: "memory", SourceClass: material.sourceClass,
		Citation: citation, Origin: material.origin, Licence: licence, IngestedAt: ingestedAt,
		Pages: ourPages, SourceTokenCount: tokenCount, TokenCountVersion: TokenCountVersion,
	}
	// The full registry mutation is recomputed under the store lock against the
	// freshly-read registry (lost-update fix): merge this event, pin the
	// consumer page set, and back-link the other cited hashes.
	merge := func(current map[string]any) (map[string]any, error) {
		newRegistry, err := MergeIngest(current, event)
		if err != nil {
			return nil, err
		}
		setConsumerPages(newRegistry, contentHash, dedupPages)
		backlinkOtherHashes(newRegistry, plan, contentHash, distilled, ingestedAt)
		return newRegistry, nil
	}

	var pageWrites []PageWrite
	for _, w := range plan.Writes {
		pageWrites = append(pageWrites, PageWrite{Filename: w.Filename, Frontmatter: w.Frontmatter, Body: w.Body})
	}
	report, err := WritePages(root, pageWrites, merge, now)
	if err != nil {
		return IngestResult{}, err
	}
	// storeOriginal runs AFTER the durable page + registry write. A failure
	// here does not un-ingest anything, so it must not be reported as total
	// failure — record it and return the successful result (iss-30).
	kept, keepErr := "", ""
	if req.KeepOriginal {
		if k, serr := storeOriginal(root, material, contentHash); serr != nil {
			keepErr = keepOriginalErrorMessage(serr)
		} else {
			kept = k
		}
	}

	status := "ingested"
	if repairing {
		status = "repaired"
	}
	return IngestResult{
		Status: status, ContentHash: contentHash, Licence: licence,
		SourceTokenCount: tokenCount, Pages: dedupPages, Citation: citation,
		KeptOriginal: kept, KeepOriginalError: keepErr,
		Linked: plan.Linked, Contradictions: plan.Contradictions,
		WriteReport: &report,
	}, nil
}

// ---------------------------------------------------------------------------
// Registry back-link helpers
// ---------------------------------------------------------------------------

func setConsumerPages(registry map[string]any, contentHash string, pages []string) {
	entry, _ := registry[contentHash].(map[string]any)
	consumers, _ := entry["consumers"].(map[string]any)
	memc, _ := consumers["memory"].(map[string]any)
	if memc != nil {
		memc["pages"] = toAnySlice(pages)
	}
}

func backlinkOtherHashes(registry map[string]any, plan WritePlan, contentHash string, distilled []DistilledPage, ingestedAt string) {
	sourceMeta := map[string]map[string]any{}
	for _, page := range distilled {
		var entries []map[string]any
		if _, ok := page.Source["class"]; ok {
			entries = []map[string]any{page.Source}
		} else if raw, ok := page.Source["sources"].([]any); ok {
			for _, e := range raw {
				if em, ok := e.(map[string]any); ok {
					entries = append(entries, em)
				}
			}
		}
		for _, se := range entries {
			if h, ok := se["source_hash"].(string); ok {
				if _, exists := sourceMeta[h]; !exists {
					sourceMeta[h] = se
				}
			}
		}
	}
	for otherHash, filenames := range plan.RegistryPages {
		if otherHash == contentHash {
			continue
		}
		entry, ok := registry[otherHash].(map[string]any)
		if !ok {
			continue
		}
		consumers, _ := entry["consumers"].(map[string]any)
		if consumers == nil {
			consumers = map[string]any{}
			entry["consumers"] = consumers
		}
		consumer, _ := consumers["memory"].(map[string]any)
		if consumer == nil {
			meta := sourceMeta[otherHash]
			class := "external_article"
			var citation map[string]any = map[string]any{}
			if meta != nil {
				if c, ok := meta["class"].(string); ok {
					class = c
				}
				if c, ok := meta["citation"].(map[string]any); ok {
					citation = deepCopyMap(c)
				}
			}
			consumer = map[string]any{
				"class": class, "citation": citation, "ingested_at": ingestedAt, "pages": []any{},
			}
			consumers["memory"] = consumer
		}
		pages := anyToStrings(consumer["pages"])
		for _, f := range filenames {
			if !contains(pages, f) {
				pages = append(pages, f)
			}
		}
		consumer["pages"] = toAnySlice(pages)
	}
}

// ---------------------------------------------------------------------------
// Registry-hit validation helpers
// ---------------------------------------------------------------------------

func pageHashSet(mem, filename string) ([]string, bool) {
	if !IsMemoryPageName(filename) {
		return nil, false
	}
	path := filepath.Join(mem, filename)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		return []string{}, true
	}
	return SourceHashes(pageSourceBlock(string(raw))), true
}

func existingPageFrontmatter(mem string) map[string]map[string]any {
	pages := map[string]map[string]any{}
	entries, err := os.ReadDir(mem)
	if err != nil {
		return pages
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || !IsMemoryPageName(e.Name()) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(mem, e.Name()))
		if err != nil {
			pages[e.Name()] = map[string]any{}
			continue
		}
		fm, err := parseFrontmatter(string(raw))
		if err != nil {
			pages[e.Name()] = map[string]any{}
			continue
		}
		pages[e.Name()] = fm
	}
	return pages
}

// ---------------------------------------------------------------------------
// Source acquisition
// ---------------------------------------------------------------------------

func isURL(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func acquireSource(source string, fetcher Fetcher, pdf PDFExtractor) (sourceMaterial, error) {
	if isURL(source) {
		var fetched FetchedSource
		var err error
		if fetcher != nil {
			fetched, err = fetcher(source)
		} else {
			fetched, err = defaultFetch(source)
		}
		if err != nil {
			var ie *IngestError
			if errors.As(err, &ie) {
				return sourceMaterial{}, err
			}
			return sourceMaterial{}, newIngestError("fetch failed for %s: %v", source, err)
		}
		return materialFromFetched(source, fetched, pdf)
	}
	return materialFromLocal(source, pdf)
}

// blockedFetchIP reports whether ip is in a range that must never be fetched
// during memory ingest: loopback (127/8, ::1), link-local (169.254/16, fe80::/10
// unicast and multicast), private (10/8, 172.16/12, 192.168/16, fc00::/7 via
// net.IP.IsPrivate), the unspecified address, and any multicast address. This is
// the SSRF guard that keeps cloud metadata endpoints (e.g. 169.254.169.254) and
// internal services out of reach.
func blockedFetchIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	// NAT64 (64:ff9b::/96) and 6to4 (2002::/16) embed an IPv4 destination in an
	// IPv6 address the checks above do not flag; a metadata/loopback/private IPv4
	// wrapped in one of these would otherwise slip through. Extract the embedded
	// v4 and re-check it (embeddedIPv4 returns nil for a v4, so no recursion loop).
	if v4 := embeddedIPv4(ip); v4 != nil {
		return blockedFetchIP(v4)
	}
	return false
}

// embeddedIPv4 returns the IPv4 address a NAT64 (64:ff9b::/96) or 6to4 (2002::/16)
// IPv6 address embeds, or nil when ip is not one of those transition forms. Scope
// is the WELL-KNOWN prefixes only: deprecated IPv4-compatible (::/96, non-routable)
// and site-specific NAT64 prefixes (RFC 8215) are out of scope — the DNS64 default
// is the well-known /96 covered here. v4-mapped ::ffff:/96 needs no extraction (the
// standard IsPrivate/IsLoopback checks already fold through To4()).
func embeddedIPv4(ip net.IP) net.IP {
	v6 := ip.To16()
	if v6 == nil || ip.To4() != nil {
		return nil // not IPv6 (a plain v4 or v4-mapped needs no extraction)
	}
	// NAT64 well-known prefix 64:ff9b::/96 → last 4 bytes are the v4.
	if v6[0] == 0x00 && v6[1] == 0x64 && v6[2] == 0xff && v6[3] == 0x9b &&
		v6[4] == 0 && v6[5] == 0 && v6[6] == 0 && v6[7] == 0 &&
		v6[8] == 0 && v6[9] == 0 && v6[10] == 0 && v6[11] == 0 {
		return net.IPv4(v6[12], v6[13], v6[14], v6[15])
	}
	// 6to4 prefix 2002::/16 → bytes 2..5 are the v4.
	if v6[0] == 0x20 && v6[1] == 0x02 {
		return net.IPv4(v6[2], v6[3], v6[4], v6[5])
	}
	return nil
}

// guardFetchHost refuses a host that is an internal/metadata name or that
// resolves to a blocked address. It runs before the initial request and on every
// redirect hop. An IP literal is checked directly (no DNS); a name is rejected
// outright when it is an *.internal / metadata name, otherwise every resolved
// address is checked.
func guardFetchHost(host string) error {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	if h == "" {
		return newIngestError("refusing to fetch a URL with no host")
	}
	if h == "metadata" || h == "metadata.google.internal" || strings.HasSuffix(h, ".internal") {
		return newIngestError("refusing to fetch internal/metadata host %q", host)
	}
	if ip := net.ParseIP(h); ip != nil {
		if blockedFetchIP(ip) {
			return newIngestError("refusing to fetch %q: address %s is link-local, loopback, private, or metadata range", host, ip)
		}
		return nil
	}
	ips, err := net.LookupIP(h)
	if err != nil {
		return newIngestError("cannot resolve host %q: %v", host, err)
	}
	for _, ip := range ips {
		if blockedFetchIP(ip) {
			return newIngestError("refusing to fetch %q: it resolves to %s (link-local, loopback, private, or metadata range)", host, ip)
		}
	}
	return nil
}

func defaultFetch(rawURL string) (FetchedSource, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return FetchedSource{}, newIngestError("fetch failed for %s: %v", rawURL, err)
	}
	if err := guardFetchHost(parsed.Hostname()); err != nil {
		return FetchedSource{}, err
	}
	// A connect-time Control hook re-checks the ACTUAL resolved IP for every
	// dialled connection, closing the DNS-rebinding gap between the name-based
	// guard above and the transport's own resolution.
	dialer := &net.Dialer{
		Timeout: fetchTimeoutSeconds * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				host = address
			}
			if ip := net.ParseIP(host); ip != nil && blockedFetchIP(ip) {
				return newIngestError("refusing to connect to %s: link-local, loopback, private, or metadata range", ip)
			}
			return nil
		},
	}
	client := &http.Client{
		Timeout:   fetchTimeoutSeconds * time.Second,
		Transport: &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return newIngestError("too many redirects fetching %s", rawURL)
			}
			return guardFetchHost(r.URL.Hostname())
		},
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return FetchedSource{}, newIngestError("fetch failed for %s: %v", rawURL, err)
	}
	req.Header.Set("User-Agent", "abcd-memory-ingest")
	resp, err := client.Do(req)
	if err != nil {
		return FetchedSource{}, newIngestError("fetch failed for %s: %v", rawURL, err)
	}
	defer resp.Body.Close()
	return readFetchedResponse(rawURL, resp)
}

// readFetchedResponse validates the HTTP status and reads the size-capped body.
// A non-2xx response (a 404/500 error page) is an ingest ERROR, not source
// content — otherwise the error page's HTML would be stored as knowledge.
func readFetchedResponse(rawURL string, resp *http.Response) (FetchedSource, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return FetchedSource{}, newIngestError("fetch failed for %s: HTTP %d %s", rawURL, resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFetchBytes+1))
	if err != nil {
		return FetchedSource{}, newIngestError("fetch failed for %s: %v", rawURL, err)
	}
	headers := map[string]string{}
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}
	final := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		final = resp.Request.URL.String()
	}
	return FetchedSource{FinalURL: final, Headers: headers, Body: body}, nil
}

func materialFromLocal(source string, pdf PDFExtractor) (sourceMaterial, error) {
	expanded := source
	// Expand only a leading "~" or "~/…" to the home dir. A "~user" form is NOT the
	// caller's home and must not be mangled into home+"user" — leave it literal.
	if expanded == "~" || strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			expanded = home + expanded[1:]
		}
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return sourceMaterial{}, newIngestError("source path is invalid: %q (%v)", source, err)
	}
	resolved := abs
	if r, err := filepath.EvalSymlinks(abs); err == nil {
		resolved = r
	}
	st, err := os.Stat(resolved)
	if err != nil {
		return sourceMaterial{}, newIngestError("source path is invalid: %q (%v)", source, err)
	}
	if !st.Mode().IsRegular() {
		return sourceMaterial{}, newIngestError("source path is not a regular file (directories, devices and symlinks-to-special are rejected): %s", resolved)
	}
	// Cap the local read the same as the URL path, so a huge local file cannot be
	// slurped whole into memory before any text/NUL sniffing.
	if st.Size() > maxFetchBytes {
		return sourceMaterial{}, newIngestError("source file exceeds the %d-byte cap: %s", maxFetchBytes, resolved)
	}
	raw, err := os.ReadFile(resolved)
	if err != nil {
		return sourceMaterial{}, newIngestError("cannot read source: %s (%v)", resolved, err)
	}
	isPDF := strings.ToLower(filepath.Ext(resolved)) == ".pdf" || strings.HasPrefix(string(raw), "%PDF-")
	if isPDF {
		text, err := extractPDFText(raw, pdf)
		if err != nil {
			return sourceMaterial{}, err
		}
		return sourceMaterial{origin: resolved, text: text, rawBytes: raw, ext: ".pdf", sourceClass: "external_pdf", title: filepath.Base(resolved)}, nil
	}
	text, err := decodeText(raw, resolved)
	if err != nil {
		return sourceMaterial{}, err
	}
	ext := safeExt(filepath.Ext(resolved))
	if ext == "" {
		ext = ".txt"
	}
	return sourceMaterial{origin: resolved, text: text, rawBytes: raw, ext: ext, sourceClass: "external_article", title: filepath.Base(resolved)}, nil
}

func materialFromFetched(rawURL string, fetched FetchedSource, pdf PDFExtractor) (sourceMaterial, error) {
	if len(fetched.Body) > maxFetchBytes {
		return sourceMaterial{}, newIngestError("fetched source exceeds the %d-byte cap: %s", maxFetchBytes, rawURL)
	}
	ctype := contentType(fetched.Headers)
	finalURL := fetched.FinalURL
	if finalURL == "" {
		finalURL = rawURL
	}
	if ctype == "application/pdf" {
		text, err := extractPDFText(fetched.Body, pdf)
		if err != nil {
			return sourceMaterial{}, err
		}
		return sourceMaterial{origin: finalURL, text: text, rawBytes: fetched.Body, headers: fetched.Headers, ext: ".pdf", sourceClass: "external_pdf", title: finalURL}, nil
	}
	if strings.HasPrefix(ctype, "text/") || textContentTypes[ctype] {
		text, err := decodeText(fetched.Body, finalURL)
		if err != nil {
			return sourceMaterial{}, err
		}
		ext := ""
		if u, err := url.Parse(finalURL); err == nil {
			ext = safeExt(filepath.Ext(u.Path))
		}
		if ext == "" {
			ext = ".txt"
		}
		return sourceMaterial{origin: finalURL, text: text, rawBytes: fetched.Body, headers: fetched.Headers, ext: ext, sourceClass: "external_article", title: finalURL}, nil
	}
	shown := ctype
	if shown == "" {
		shown = "(missing)"
	}
	return sourceMaterial{}, newIngestError("non-text content-type %q rejected for %s; nothing written", shown, finalURL)
}

func extractPDFText(data []byte, pdf PDFExtractor) (string, error) {
	if pdf == nil {
		return "", newIngestError("PDF extraction unavailable: no PDF text-extraction dependency is installed (supply a PDFExtractor)")
	}
	text, err := pdf(data)
	if err != nil {
		return "", newIngestError("PDF extraction failed: %v", err)
	}
	if strings.TrimSpace(text) == "" {
		return "", newIngestError("PDF has no extractable text; nothing to ingest")
	}
	return text, nil
}

func decodeText(data []byte, what string) (string, error) {
	for _, b := range data {
		if b == 0 {
			return "", newIngestError("binary source rejected: %s contains NUL bytes and no text-extraction path applies", what)
		}
	}
	if !utf8.Valid(data) {
		return "", newIngestError("binary source rejected: %s is not decodable text", what)
	}
	return string(data), nil
}

func contentType(headers map[string]string) string {
	for k, v := range headers {
		if strings.ToLower(k) == "content-type" {
			return strings.ToLower(strings.TrimSpace(strings.SplitN(v, ";", 2)[0]))
		}
	}
	return ""
}

func safeExt(ext string) string {
	ext = strings.ToLower(ext)
	if extRe.MatchString(ext) {
		return ext
	}
	return ""
}

// ---------------------------------------------------------------------------
// Original storage (--keep-original)
// ---------------------------------------------------------------------------

// sourcesRelPath is the repo-relative location of the kept-originals store,
// used in user-facing errors so no absolute path leaks into rendered output.
var sourcesRelPath = filepath.Join(".abcd", "memory", "sources")

// keepOriginalErrorMessage renders a --keep-original failure without leaking the
// absolute sources path: filesystem errors embed the full path(s), so report
// only their bare cause against the repo-relative store location (iss-30). Both
// *PathError (Lstat/MkdirAll/OpenFile/Write/Sync) and *LinkError (Rename, which
// carries TWO absolute paths) are stripped; the only other storeOriginal error
// already names the repo-relative sourcesRelPath.
func keepOriginalErrorMessage(err error) string {
	if pe := (*os.PathError)(nil); errors.As(err, &pe) {
		return fmt.Sprintf("could not store original under %s: %s", sourcesRelPath, pe.Err.Error())
	}
	if le := (*os.LinkError)(nil); errors.As(err, &le) {
		return fmt.Sprintf("could not store original under %s: %s", sourcesRelPath, le.Err.Error())
	}
	return err.Error()
}

func storeOriginal(repoRoot string, material sourceMaterial, contentHash string) (string, error) {
	sourcesDir := filepath.Join(Dir(repoRoot), "sources")
	if fi, err := os.Lstat(sourcesDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return "", newIngestError("sources dir is a symlink or non-directory: %s", sourcesRelPath)
	}
	target := filepath.Join(sourcesDir, contentHash+material.ext)
	tmp := target + "." + itoa(os.Getpid()) + ".memtmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(material.rawBytes); err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmp)
		return "", err
	}
	f.Close()
	if err := os.Rename(tmp, target); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return filepath.Join(".abcd", "memory", "sources", contentHash+material.ext), nil
}

func dedupStrings(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
