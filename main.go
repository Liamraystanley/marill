// Author: Liam Stanley <me@liamstanley.io>
// Docs: https://marill.liam.sh/
// Repo: https://github.com/Liamraystanley/marill
//
//       O
//    o 0  o        [ Marill -- Automated site testing utility ]
//       O      ___      ___       __        _______    __    ___      ___
//     o       |"  \    /"  |     /""\      /"      \  |" \  |"  |    |"  |
//    [  ]      \   \  //   |    /    \    |:        | ||  | ||  |    ||  |
//    / O\      /\\  \/.    |   /' /\  \   |_____/   ) |:  | |:  |    |:  |
//   / o  \    |: \.        |  //  __'  \   //      /  |.  |  \  |___  \  |___
//  / O  o \   |.  \    /:  | /   /  \\  \ |:  __   \  /\  |\( \_|:  \( \_|:  \
// [________]  |___|\__/|___|(___/    \___)|__|  \___)(__\_|_)\_______)\_______)
//

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Liamraystanley/marill/domfinder"
	"github.com/Liamraystanley/marill/scraper"
	"github.com/Liamraystanley/marill/utils"
	"github.com/urfave/cli"
)

// these /SHOULD/ be defined during the make process. not always however.
var version, commithash, compiledate = "", "", ""

const motd = `
{magenta}      {lightgreen}O{magenta}     {yellow}     [ Marill -- Automated site testing utility ]
{magenta}   {lightgreen}o{magenta} {lightgreen}0{magenta}  {lightgreen}o{magenta}   {lightyellow}             %4s, rev %s
{magenta}      {lightgreen}O{magenta}     {lightblue} ___      ___       __        _______    __    ___      ___
{magenta}    {lightgreen}o{magenta}       {lightblue}|"  \    /"  |     /""\      /"      \  |" \  |"  |    |"  |
{magenta}   [  ]     {lightblue} \   \  //   |    /    \    |:        | ||  | ||  |    ||  |
{magenta}   / {lightmagenta}O{magenta}\     {lightblue} /\\  \/.    |   /' /\  \   |_____/   ) |:  | |:  |    |:  |
{magenta}  / {lightmagenta}o{magenta}  \    {lightblue}|: \.        |  //  __'  \   //      /  |.  |  \  |___  \  |___
{magenta} / {lightmagenta}O{magenta}  {lightmagenta}o{magenta} \   {lightblue}|.  \    /:  | /   /  \\  \ |:  __   \  /\  |\( \_|:  \( \_|:  \
{magenta}[________]  {lightblue}|___|\__/|___|(___/    \___)|__|  \___)(__\_|_)\_______)\_______)

`

// outputConfig handles what the user sees (stdout, debugging, logs, etc)
type outputConfig struct {
	noColors   bool   // don't print colors to stdout
	noBanner   bool   // don't print the app banner
	printDebug bool   // print debugging information
	ignoreStd  bool   // ignore regular stdout (human-formatted)
	logFile    string // optional log file to dump debugging info
}

// scanConfig handles how and what is scanned/crawled
type scanConfig struct {
	threads    int           // number of threads to run the scanner in
	manualList string        // list of manually supplied domains
	recursive  bool          // recursively pull resources, and their assets
	delay      time.Duration // delay for the stasrt of each resource crawl

	// domain filter related
	ignoreHTTP   bool   // ignore http://
	ignoreHTTPS  bool   // ignore https://
	ignoreRemote bool   // ignore resources where the domain is using remote ip
	ignoreMatch  string // glob match of domains to blacklist
	matchOnly    string // glob match of domains to whitelist

	// test related
	ignoreTest     string  // glob match of tests to blacklist
	matchTest      string  // glob match of tests to whitelist
	minScore       float64 // minimum score before a resource is considered "failed"
	testsFromURL   string  // load tests from a remote url
	testsFromPath  string  // load tests from a specified path
	ignoreStdTests bool    // don't execute standard builtin tests
}

// appConfig handles what the app does (scans/crawls, printing data, some other task, etc)
type appConfig struct {
	printUrls          bool
	printTests         bool
	printTestsExtended bool
}

// config is a wrapper for all the other configs to put them in one place
type config struct {
	app  appConfig
	scan scanConfig
	out  outputConfig
}

var conf config
var out = Output{}

// statsLoop prints out memory/load/runtime statistics to debug output
func statsLoop(done <-chan struct{}) {
	mem := &runtime.MemStats{}
	var numRoutines, numCPU int
	var load5, load10, load15 float32

	for {
		select {
		case <-done:
			return
		default:
			runtime.ReadMemStats(mem)
			numRoutines = runtime.NumGoroutine()
			numCPU = runtime.NumCPU()

			if contents, err := ioutil.ReadFile("/proc/loadavg"); err == nil {
				fmt.Sscanf(string(contents), "%f %f %f %*s %*d", &load5, &load10, &load15)
			}

			logger.Printf(
				"allocated mem: %dM, sys: %dM, routines: %d, cores: %d load5: %.2f load10: %.2f load15: %.2f",
				mem.Alloc/1024/1024, mem.Sys/1024/1024, numRoutines, numCPU, load5, load10, load15)

			time.Sleep(2 * time.Second)
		}
	}
}

// numThreads calculates the number of threads to use
func numThreads() {
	// use runtime.NumCPU for judgement call
	if conf.scan.threads < 1 {
		if runtime.NumCPU() >= 2 {
			conf.scan.threads = runtime.NumCPU() / 2
		} else {
			conf.scan.threads = 1
		}
	} else if conf.scan.threads > runtime.NumCPU() {
		logger.Printf("warning: %d threads specified, which is more than the amount of cores", conf.scan.threads)
		out.Printf("{yellow}warning: %d threads specified, which is more than the amount of cores on the server!{c}\n", conf.scan.threads)

		// set it to the amount of cores on the server. go will do this regardless, so.
		conf.scan.threads = runtime.NumCPU()
		logger.Printf("limiting number of threads to %d", conf.scan.threads)
		out.Printf("limiting number of threads to %d\n", conf.scan.threads)
	}

	logger.Printf("using %d cores (max %d)", conf.scan.threads, runtime.NumCPU())

	return
}

// reManualDomain can match the following:
// (DOMAIN|URL):IP:PORT
// (DOMAIN|URL):IP
// (DOMAIN|URL):PORT
// (DOMAIN|URL)
var reManualDomain = regexp.MustCompile(`^(?P<domain>(?:[A-Za-z0-9_.-]{2,350}\.[A-Za-z0-9]{2,63})|https?://[A-Za-z0-9_.-]{2,350}\.[A-Za-z0-9]{2,63}[!-~]+?)(?::(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}))?(?::(?P<port>\d{2,5}))?$`)
var reSpaces = regexp.MustCompile(`[\t\n\v\f\r ]+`)

/// parseManualList parses the list of domains specified from --domains
func parseManualList() (domlist []*scraper.Domain, err error) {
	input := strings.Split(reSpaces.ReplaceAllString(conf.scan.manualList, " "), " ")

	for _, item := range input {
		item = strings.TrimSuffix(strings.TrimPrefix(item, " "), " ")
		if item == "" {
			continue
		}

		results := reManualDomain.FindStringSubmatch(item)
		if len(results) != 4 {
			return nil, NewErr{Code: ErrBadDomainFlag, value: item}
		}

		domain, ip, port := results[1], results[2], results[3]

		if domain == "" {
			return nil, NewErr{Code: ErrBadDomainFlag, value: item}
		}

		uri, err := utils.IsDomainURL(domain, port)
		if err != nil {
			return nil, NewErr{Code: ErrBadDomainFlag, deepErr: err}
		}

		domlist = append(domlist, &scraper.Domain{
			URL: uri,
			IP:  ip,
		})
	}

	return domlist, nil
}

// printUrls prints the urls that /would/ be scanned, if we were to start crawling
func printUrls() {
	printBanner()

	if conf.scan.manualList != "" {
		domains, err := parseManualList()
		if err != nil {
			out.Fatal(NewErr{Code: ErrDomainFlag, deepErr: err})
		}

		for _, domain := range domains {
			out.Printf("{blue}%-40s{c} {green}%s{c}\n", domain.URL, domain.IP)
		}
	} else {
		finder := &domfinder.Finder{Log: logger}
		if err := finder.GetWebservers(); err != nil {
			out.Fatal(NewErr{Code: ErrProcList, deepErr: err})
		}

		if err := finder.GetDomains(); err != nil {
			out.Fatal(NewErr{Code: ErrGetDomains, deepErr: err})
		}

		finder.Filter(domfinder.DomainFilter{
			IgnoreHTTP:  conf.scan.ignoreHTTP,
			IgnoreHTTPS: conf.scan.ignoreHTTPS,
			IgnoreMatch: conf.scan.ignoreMatch,
			MatchOnly:   conf.scan.matchOnly,
		})

		if len(finder.Domains) == 0 {
			out.Fatal(NewErr{Code: ErrNoDomainsFound})
		}

		for _, domain := range finder.Domains {
			out.Printf("{blue}%-40s{c} {green}%s{c}\n", domain.URL, domain.IP)
		}
	}
}

// listTests lists all loaded tests, based on supplied args to Marill
func listTests() {
	printBanner()

	tests := genTests()

	out.Printf("{lightgreen}%d{c} total tests found:\n", len(tests))

	for _, test := range tests {
		weightID := "-"
		if !test.Bad {
			weightID = "+"
		}

		out.Printf("{lightblue}n:{c} %-25s {lightblue}t:{c} %-13s {lightblue}w:{c} %s%-6.2f {lightblue}o:{c} %s\n", test.Name, test.Type, weightID, test.Weight, test.Origin)

		if conf.app.printTestsExtended {
			if len(test.MatchRegex) > 0 {
				out.Printf("    - {cyan}regex{c}: {yellow}[{c}%s{yellow}]{c}\n", strings.Join(test.MatchRegex, "{yellow}]{c}, {yellow}[{c}"))
			}

			if len(test.Match) > 0 {
				out.Printf("    -  {cyan}glob{c}: {yellow}[{c}%s{yellow}]{c}\n", strings.Join(test.Match, "{yellow}]{c}, {yellow}[{c}"))
			}

			out.Println("")
		}
	}
}

func printBanner() {
	if len(version) != 0 && len(commithash) != 0 {
		logger.Printf("marill: version:%s revision:%s\n", version, commithash)
		if conf.out.noBanner {
			out.Printf("{bold}{blue}marill version: %s (rev %s)\n", version, commithash)
		} else {
			out.Printf(motd, version, commithash)
		}
	} else {
		out.Println("{bold}{blue}Running marill (unknown version){c}")
	}
}

func run() {
	printBanner()

	// fetch the tests ahead of time to ensure there are no syntax errors or anything
	tests := genTests()

	crawler := &scraper.Crawler{Log: logger}
	var err error

	if conf.scan.manualList != "" {
		logger.Println("manually supplied url list")
		crawler.Cnf.Domains, err = parseManualList()
		if err != nil {
			out.Fatal(NewErr{Code: ErrDomainFlag, deepErr: err})
		}
	} else {
		logger.Println("checking for running webservers")

		finder := &domfinder.Finder{Log: logger}
		if err := finder.GetWebservers(); err != nil {
			out.Fatal(NewErr{Code: ErrProcList, deepErr: err})
		}

		if outlist := ""; len(finder.Procs) > 0 {
			for _, proc := range finder.Procs {
				outlist += fmt.Sprintf("[%s:%s] ", proc.Name, proc.PID)
			}
			logger.Printf("found %d procs matching a webserver: %s", len(finder.Procs), outlist)
			out.Printf("found %d procs matching a webserver\n", len(finder.Procs))
		}

		// start crawling for domains
		if err := finder.GetDomains(); err != nil {
			out.Fatal(NewErr{Code: ErrGetDomains, deepErr: err})
		}

		finder.Filter(domfinder.DomainFilter{
			IgnoreHTTP:  conf.scan.ignoreHTTP,
			IgnoreHTTPS: conf.scan.ignoreHTTPS,
			IgnoreMatch: conf.scan.ignoreMatch,
			MatchOnly:   conf.scan.matchOnly,
		})

		if len(finder.Domains) == 0 {
			out.Fatal(NewErr{Code: ErrNoDomainsFound})
		}

		logger.Printf("found %d domains on webserver %s (exe: %s, pid: %s)", len(finder.Domains), finder.MainProc.Name, finder.MainProc.Exe, finder.MainProc.PID)

		for _, domain := range finder.Domains {
			crawler.Cnf.Domains = append(crawler.Cnf.Domains, &scraper.Domain{URL: domain.URL, IP: domain.IP})
		}
	}

	crawler.Cnf.Recursive = conf.scan.recursive
	crawler.Cnf.NoRemote = conf.scan.ignoreRemote
	crawler.Cnf.Delay = conf.scan.delay

	logger.Printf("starting crawler...")
	out.Printf("starting scan on %d domains\n", len(crawler.Cnf.Domains))
	crawler.Crawl()
	out.Println("{lightgreen}scan complete{c}")

	out.Println("{lightblue}starting tests{c}")
	testResults := checkTests(crawler.Results, tests)

	for _, res := range testResults {
		if res.Domain.Error != nil {
			out.Printf("{red}[FAILURE]{c} %5.1f/10 [code: ---] [%15s] [{cyan}  0 resources{c}] [{green}     0ms{c}] %s ({red}%s{c})\n", res.Score, res.Domain.Request.IP, res.Domain.Request.URL, res.Domain.Error)
		} else {
			url := res.Domain.Resource.Response.URL.String()
			if url != res.Domain.Request.URL {
				url = fmt.Sprintf("%s (result: %s)", res.Domain.Request.URL, url)
			}

			out.Printf("{green}[SUCCESS]{c} %5.1f/10 [code: {yellow}%d{c}] [%15s] [{cyan}%3d resources{c}] [{green}%6dms{c}] %s\n", res.Score, res.Domain.Resource.Response.Code, res.Domain.Request.IP, len(res.Domain.Resources), res.Domain.Resource.Time.Milli, url)
		}
	}
}

func main() {
	defer closeLogger() // ensure we're cleaning up the logger if there is one

	cli.VersionPrinter = func(c *cli.Context) {
		if version != "" && commithash != "" && compiledate != "" {
			fmt.Printf("version %s, revision %s (%s)\n", version, commithash, compiledate)
		} else if commithash != "" && compiledate != "" {
			fmt.Printf("revision %s (%s)\n", commithash, compiledate)
		} else if version != "" {
			fmt.Printf("version %s\n", version)
		} else {
			fmt.Println("version unknown")
		}
	}

	app := cli.NewApp()

	app.Name = "marill"

	if version != "" && commithash != "" {
		app.Version = fmt.Sprintf("%s, git revision %s", version, commithash)
	} else if version != "" {
		app.Version = version
	} else if commithash != "" {
		app.Version = "git revision " + commithash
	}

	// needed for stats look
	done := make(chan struct{}, 1)

	app.Before = func(c *cli.Context) error {
		// initialize the logger
		initLogger()

		// initialize the max amount of threads to use
		numThreads()

		// initialize the stats data
		go statsLoop(done)

		return nil
	}

	app.After = func(c *cli.Context) error {
		// close the stats data goroutine when we're complete.
		done <- struct{}{}

		return nil
	}

	app.Flags = []cli.Flag{
		// output style flags
		cli.BoolFlag{
			Name:        "d, debug",
			Usage:       "Print debugging information to stdout",
			Destination: &conf.out.printDebug,
		},
		cli.BoolFlag{
			Name:        "q, quiet",
			Usage:       "Do not print regular stdout messages",
			Destination: &conf.out.ignoreStd,
		},
		cli.BoolFlag{
			Name:        "no-color",
			Usage:       "Do not print with color",
			Destination: &conf.out.noColors,
		},
		cli.BoolFlag{
			Name:        "no-banner",
			Usage:       "Do not print the colorful banner",
			Destination: &conf.out.noBanner,
		},
		cli.StringFlag{
			Name:        "log-file",
			Usage:       "Log debugging information to `logfile`",
			Destination: &conf.out.logFile,
		},

		// app related
		cli.BoolFlag{
			Name:        "urls",
			Usage:       "Print the list of urls as if they were going to be scanned",
			Destination: &conf.app.printUrls,
		},
		cli.BoolFlag{
			Name:        "tests",
			Usage:       "Print the list of tests that are loaded and would be used",
			Destination: &conf.app.printTests,
		},
		cli.BoolFlag{
			Name:        "tests-extended",
			Usage:       "Same as --tests, with extra information",
			Destination: &conf.app.printTestsExtended,
		},

		// scan configuration
		cli.IntFlag{
			Name:        "threads",
			Usage:       "Use `n` threads to fetch data (0 defaults to server cores/2)",
			Destination: &conf.scan.threads,
		},
		cli.DurationFlag{
			Name:        "delay",
			Usage:       "Delay `DURATION` before each resource is crawled (e.g. 5s, 1m, 100ms)",
			Destination: &conf.scan.delay,
		},
		cli.StringFlag{
			Name:        "domains",
			Usage:       "Manually specify list of domains to scan in form: `DOMAIN:IP ...`, or DOMAIN:IP:PORT",
			Destination: &conf.scan.manualList,
		},
		cli.Float64Flag{
			Name:        "min-score",
			Usage:       "Minimium score for domain",
			Value:       8.0,
			Destination: &conf.scan.minScore,
		},
		cli.BoolFlag{
			Name:        "r, recursive",
			Usage:       "Check all assets (css/js/images) for each page, recursively",
			Destination: &conf.scan.recursive,
		},

		// domain filtering
		cli.BoolFlag{
			Name:        "ignore-http",
			Usage:       "Ignore http-based URLs during domain search",
			Destination: &conf.scan.ignoreHTTP,
		},
		cli.BoolFlag{
			Name:        "ignore-https",
			Usage:       "Ignore https-based URLs during domain search",
			Destination: &conf.scan.ignoreHTTPS,
		},
		cli.BoolFlag{
			Name:        "ignore-remote",
			Usage:       "Ignore all resources that resolve to a remote IP (use with --recursive)",
			Destination: &conf.scan.ignoreRemote,
		},
		cli.StringFlag{
			Name:        "domain-ignore",
			Usage:       "Ignore URLS during domain search that match `GLOB`, pipe separated list",
			Destination: &conf.scan.ignoreMatch,
		},
		cli.StringFlag{
			Name:        "domain-match",
			Usage:       "Allow URLS during domain search that match `GLOB`, pipe separated list",
			Destination: &conf.scan.matchOnly,
		},

		// test filtering
		cli.StringFlag{
			Name:        "test-ignore",
			Usage:       "Ignore tests that match `GLOB`, pipe separated list",
			Destination: &conf.scan.ignoreTest,
		},
		cli.StringFlag{
			Name:        "test-match",
			Usage:       "Allow tests that match `GLOB`, pipe separated list",
			Destination: &conf.scan.matchTest,
		},
		cli.StringFlag{
			Name:        "tests-url",
			Usage:       "Import tests from a specified `URL`",
			Destination: &conf.scan.testsFromURL,
		},
		cli.StringFlag{
			Name:        "tests-path",
			Usage:       "Import tests from a specified file-system `PATH`",
			Destination: &conf.scan.testsFromPath,
		},
		cli.BoolFlag{
			Name:        "ignore-std-tests",
			Usage:       "Ignores all built-in tests (useful with --tests-url)",
			Destination: &conf.scan.ignoreStdTests,
		},
	}

	app.Authors = []cli.Author{
		{
			Name:  "Liam Stanley",
			Email: "me@liamstanley.io",
		},
	}
	app.Copyright = "(c) 2016 Liam Stanley"
	app.Compiled = time.Now()
	app.Usage = "Automated website testing utility"
	app.Action = func(c *cli.Context) error {
		if conf.app.printUrls {
			printUrls()
			os.Exit(0)
		} else if conf.app.printTests || conf.app.printTestsExtended {
			listTests()
			os.Exit(0)
		}

		run()

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(NewErr{Code: ErrInstantiateApp, deepErr: err})
	}
}
