import subprocess
import csv
import sys
from argparse import ArgumentParser
import re
import os

def write_coverage(coverage_map, filename):
    with open(filename, "w") as csvfile:
        coverage_writer = csv.writer(csvfile, delimiter=";")
        for key in coverage_map.keys():
            coverage_writer.writerow([key, coverage_map[key]])

def read_coverage(filename):
    coverage_map = {}
    with open(filename, "r") as csvfile:
        coverage_reader = csv.reader(csvfile, delimiter=";")
        for row in coverage_reader:
                coverage_map[row[0]] = float(row[1])
    return coverage_map


scriptdir = os.path.dirname(os.path.realpath(__file__))

parser = ArgumentParser()
parser.add_argument("-w", "--write-new-ref", dest="write_new_ref", default=False,
                    help="write new reference file")


args = parser.parse_args()
filename = os.path.join(scriptdir,"coverage.csv")
reference_coverage_map = read_coverage(filename)
package = "github.infra.hana.ondemand.com/istio/istio-broker"
command = "go test -cover %s/..." % package
print(command)

test = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE).stdout.read()
test_output = test.decode("utf-8")
print( test_output)
#should look like this:
#ok  	github.infra.hana.ondemand.com/istio/istio-broker	(cached)	coverage: 0.0% of statements [no tests to run]
#ok  	github.infra.hana.ondemand.com/istio/istio-broker/integration	(cached)	coverage: 0.0% of statements

coverage_map = {}

test_lines = test_output.split("\n")
for test_line in test_lines:
    if len(test_line) > 0:
        test_result = re.split(r'\s+', test_line)
        coverage = test_result[-3]
        if "%" not in coverage:
            coverage = test_result[-7]
        coverage = float(coverage[:-1])
        path = test_result[1]
        coverage_map[path] = coverage

if args.write_new_ref:
        print("Writing a new reference.")
        write_coverage(coverage_map, filename)
        exit (0)

coverage_better = False
coverage_worse = False

for key in coverage_map.keys():
        try:
                if coverage_map[key] < reference_coverage_map[key]:
                        print ("{}: {}% < {}%".format(key, coverage_map[key], reference_coverage_map[key]))
                        coverage_worse = True
                elif coverage_map[key] > reference_coverage_map[key]:
                        coverage_better =True
                        
        except KeyError:
                print("Can't compare for this path is new: {}".format(key)) 
                coverage_better = True

if coverage_worse:
        print("Coverage got worse!")
        exit (1)

if coverage_better:
        print("Coverage got bestter. Writing new reference.")
        write_coverage(coverage_map, filename)
                
        os.chdir(scriptdir)
        subprocess.call(["git", "commit", "-a", "-m", "`coverage improved so reference is saved`"])
        subprocess.call(["git", "push"])
        