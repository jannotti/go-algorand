# End to end tests

This directory contains the category of tests which we like to call "end to end". Primarily they consist of tests which first start a private network and then run a series of commands against that network.

These tests have grown since the project started and we have a number of different frameworks. There are a number of different tests, frameworks and tools in this directory.


# Directories
## Tests / Test Frameworks

* scripts - shell scripted integration test framework.
* e2e-go - tests that can be run with the `go test`.
* framework - functions and utilities used by the e2e-go tests.
* release-testing - a spot for specific release tests, see README files in subdirectories.
* muleCI - scripts run tests on a Jenkins server with mule
* packages - test that algod can be packaged on different docker environments.
* platform - test the algod amd64 package compatibility across different distributions.

## Tools / Data
* commandandcontrol - a remote control tool for algod. It allows you to manage many instances across many nodes.
* netperf-go - tools for semi-automated performance tests.
* testdata - datasets used by other tools not included in this repository.

# Scripts

Entry point to our integration test framework, including the e2e-go tests.

Must run from the root project directory, `./test/scripts/e2e.sh`

## scripts/e2e_client_runner.py and scripts/e2e_subs/

These tests are shell scripts which all run in parallel against a single private network.

Each script is provided with a wallet which contains a large supply of algos to use during the test.
```
usage: e2e_client_runner.py [-h] [--keep-temps] [--timeout TIMEOUT] [--verbose] [--version Future|vXX] [scripts [scripts ...]]

positional arguments:
  scripts            scripts to run

optional arguments:
  -h, --help             show this help message and exit
  --keep-temps           if set, keep all the test files
  --timeout TIMEOUT      integer seconds to wait for the scripts to run
  --verbose
  --version Future|vXX   selects the network template file
```

To run a specific test, run e2e.sh with -i interactive flag, and follow the instructions:
```
test/scripts/e2e.sh -i
```

Or drive the runner yourself, which is quicker when iterating on one script. It
needs `goal`, `algod` and `kmd` built from your branch, and a virtualenv with
the Python SDK:
```bash
make install                 # puts binaries in $(go env GOPATH)/bin; `goal -v` prints the branch
python3 -m venv /tmp/ve && /tmp/ve/bin/pip3 install 'py-algorand-sdk==2.*'

PATH=$(go env GOPATH)/bin:$PATH /tmp/ve/bin/python3 \
    ./test/scripts/e2e_client_runner.py --unsafe_scrypt \
    ./test/scripts/e2e_subs/lsig-salted.sh ./test/scripts/e2e_subs/pq-falcon.sh
```
`--unsafe_scrypt` makes kmd much faster and is what CI uses. Each script brings
up its own network and takes roughly half a minute. Pass `--keep-temps` to keep
the network directory when a script fails, and note that `--version` defaults to
`Future`, so vFuture features work without extra flags.

Tests in the `e2e_subs/serial` directory are executed serially instead of in parallel. This should only be used when absolutely necessary.