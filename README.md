![Marill -- Automated Site Testing Utility](https://i.imgur.com/HYZ3biA.png)
<p align="center">Marill -- Automated site testing utility.</p>

[![Build Status](https://travis-ci.org/Liamraystanley/marill.svg?branch=master)](https://travis-ci.org/Liamraystanley/marill)
[![GitHub Issues](https://img.shields.io/github/issues/Liamraystanley/marill.svg)](https://github.com/Liamraystanley/marill/issues)
[![Project Status](https://img.shields.io/badge/status-alpha-red.svg)](https://github.com/Liamraystanley/marill/commits/master)
[![Codebeat Badge](https://codebeat.co/badges/4653f785-83ec-4b21-bf0c-b519b20c89d6)](https://codebeat.co/projects/github-com-liamraystanley-marill)
[![Go Report Card](https://goreportcard.com/badge/github.com/Liamraystanley/marill)](https://goreportcard.com/report/github.com/Liamraystanley/marill)

## Project Status:
At this stage, things are still in alpha/likely going to change quite a bit. This includes code, exported functions/tools, cli args, etc.

   * [See here](https://github.com/Liamraystanley/marill/projects/1) for what is being worked on/in my todo list for the first beta release.
   * [See here](https://github.com/Liamraystanley/marill/projects/2) for what is being worked on/in my todo list for the first major release.
   * Head over to [release.liam.sh/marill](https://release.liam.sh/marill/?sort=time&order=desc) to get some testing bundled binaries, if your'e daring and willing to help test my latest pushes.

## Building:
Marill supports building on 1.3+ (or even possibly older), however it is recommended to build on the latest go release. Note that you will not be able to use the Makefile to compile Marill if you are trying to build on go 1.4 or lower. You will need to manually compile it, due to limitations with ldflag support.

```
$ git clone https://github.com/Liamraystanley/marill.git
$ cd marill
$ make
```

To run unit tests, then compile, simply run:

```
$ make test all
```

## Usage:
This is very likely to change quite a bit until we're out of beta. Please use wisely.

```
$ marill --help
NAME:
   marill - Automated website testing utility

USAGE:
   marill [global options] command [command options] [arguments...]
   
VERSION:
   git revision XXXXXX
   
AUTHOR(S):
   Liam Stanley <user@domain.com> 
   
GLOBAL OPTIONS:
   -d, --debug              Print debugging information to stdout
   -q, --quiet              Do not print regular stdout messages
   --no-color               Do not print with color
   --no-banner              Do not print the colorful banner
   --exit-on-fail           Send exit code 1 if any domains fail tests
   --log-file logfile       Log debugging information to logfile
   --urls                   Print the list of urls as if they were going to be scanned
   --tests                  Print the list of tests that are loaded and would be used
   --tests-extended         Same as --tests, with extra information
   --threads n              Use n threads to fetch data (0 defaults to server cores/2) (default: 0)
   --delay DURATION         Delay DURATION before each resource is crawled (e.g. 5s, 1m, 100ms) (default: 0s)
   --domains DOMAIN:IP ...  Manually specify list of domains to scan in form: DOMAIN:IP ..., or DOMAIN:IP:PORT
   --min-score value        Minimium score for domain (default: 8)
   -r, --recursive          Check all assets (css/js/images) for each page, recursively
   --ignore-http            Ignore http-based URLs during domain search
   --ignore-https           Ignore https-based URLs during domain search
   --ignore-remote          Ignore all resources that resolve to a remote IP (use with --recursive)
   --domain-ignore GLOB     Ignore URLS during domain search that match GLOB, pipe separated list
   --domain-match GLOB      Allow URLS during domain search that match GLOB, pipe separated list
   --test-ignore GLOB       Ignore tests that match GLOB, pipe separated list
   --test-match GLOB        Allow tests that match GLOB, pipe separated list
   --tests-url URL          Import tests from a specified URL
   --tests-path PATH        Import tests from a specified file-system PATH
   --ignore-std-tests       Ignores all built-in tests (useful with --tests-url)
   --help, -h               show help
   --version, -v            print the version
   
COPYRIGHT:
   (c) 2016 Liam Stanley (see https://git.io/vPvUp)
```

## License:

    LICENSE: The MIT License (MIT)
    Copyright (c) 2016 Liam Stanley <me@liamstanley.io>

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:
    
    The above copyright notice and this permission notice shall be included in
    all copies or substantial portions of the Software.
    
    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE.
