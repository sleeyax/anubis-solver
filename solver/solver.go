package solver

import (
	"anubis-solver/polyfills/fetch"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strings"
	"time"
)

const testedVersion = "v1.16.0-36-ga14f917"

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

type Solver struct {
	URL    string
	client tlsclient.HttpClient
}

func New(url string) (*Solver, error) {
	jar := tlsclient.NewCookieJar()
	options := []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(30),
		tlsclient.WithClientProfile(profiles.Chrome_120),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithCookieJar(jar),
		tlsclient.WithNotFollowRedirects(),
	}

	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	return &Solver{URL: url, client: client}, nil
}

func (s *Solver) buildHeaders() http.Header {
	return http.Header{
		"host":                      {"anubis.techaro.lol"},
		"sec-ch-ua":                 {`"Not:A-Brand";v="24", "Chromium";v="134"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {"Linux"},
		"accept-language":           {"en-US,en;q=0.9"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {fetch.DefaultUserAgent},
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,pensive/avif,pensive/webp,pensive/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-user":            {"?1"},
		"sec-fetch-dest":            {"document"},
		"accept-encoding":           {"gzip, deflate, br"},
		"priority":                  {"u=0, i"},
		"connection":                {"keep-alive"},
		http.HeaderOrderKey: {
			"host",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"accept-language",
			"upgrade-insecure-requests",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
			"priority",
			"connection",
		},
	}
}

func (s *Solver) Solve() error {
	req, err := http.NewRequest(http.MethodGet, s.URL, nil)
	if err != nil {
		return err
	}
	req.Header = s.buildHeaders()

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status code: " + res.Status)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	version := s.formatVersion(doc.Find("#anubis_version").Get(0).FirstChild.Data)
	log.Info().Msg(fmt.Sprintf("found version: %s", version))
	if version != testedVersion {
		log.Warn().Msg("this version has not been tested yet, it may not work correctly")
	}

	challengeJson := doc.Find("#anubis_challenge").Get(0).FirstChild.Data
	log.Debug().Str("json", challengeJson).Msg("found challenge")
	challenge := &AnubisChallenge{}
	if err = json.Unmarshal([]byte(challengeJson), challenge); err != nil {
		return err
	}

	pensive, exists := doc.Find("#image").Attr("src")
	if !exists {
		return errors.New("pensive image not found")
	}
	log.Debug().Str("pensive", pensive).Msg("found pensive image")

	happy, exists := doc.Find("img[style='display:none;']").Attr("src")
	if !exists {
		return errors.New("happy image not found")
	}
	log.Debug().Str("happy", happy).Msg("found happy image")

	scriptPath, exists := doc.Find("script[type='module']").Attr("src")
	if !exists {
		return errors.New("script not found")
	}
	log.Debug().Str("script", scriptPath).Msg("found script")

	// Fetch the script source code.
	scriptUrl := s.URL + strings.TrimLeft(scriptPath, "/")
	req, err = http.NewRequest(http.MethodGet, scriptUrl, nil)
	if err != nil {
		return err
	}
	req.Header = s.buildHeaders()
	res, err = s.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status code while fetching script source: " + res.Status)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Solve the POW challenge.
	log.Info().Msg("solving challenge...")
	start := time.Now()
	submissionUrl, err := SolveChallenge(s.URL, string(body), version, challengeJson)
	if err != nil {
		return fmt.Errorf("failed to solve challenge: %w", err)
	}
	end := time.Now()
	log.Info().Str("url", submissionUrl).Msg(fmt.Sprintf("solved challenge in %s", end.Sub(start)))

	// Submit the solution and retrieve the cookie.
	req, err = http.NewRequest(http.MethodGet, submissionUrl, nil)
	if err != nil {
		return err
	}
	req.Header = s.buildHeaders()
	res, err = s.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusFound {
		return errors.New("unexpected status code while submitting solution: " + res.Status)
	}
	defer res.Body.Close()
	cookie := res.Header.Get("set-cookie")
	log.Info().Str("cookie", cookie).Msg("retrieved cookie")

	return nil
}

func (s *Solver) formatVersion(version string) string {
	return strings.Trim(strings.TrimSpace(version), `""`)
}
