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


parser = ArgumentParser()
parser.add_argument("-w", "--write-new-ref", dest="write_new_ref", default=False,
                    help="write new reference file")

parser.add_argument("-p", "--push-better-ref", dest="push_better_ref", default=False,
                    help="push new reference file if coverage got better")

parser.add_argument("-g", "--go-package", dest="go_package", default="github.com/Peripli/istio-broker-proxy",
                    help="push new reference file if coverage got better")

args = parser.parse_args()
go_path = os.environ['GOPATH']
src_path = os.path.join(go_path, "src", args.go_package)

filename = os.path.join(src_path,"coverage.csv")
reference_coverage_map = read_coverage(filename)
package = args.go_package
command = "go test -cover %s/..." % package
print(command)

try:
    test = subprocess.check_output(command, shell=True)
except subprocess.CalledProcessError as e:
    print(e.output.decode("unicode_escape"))
    raise

test_output = test.decode("utf-8")
print( test_output)
#should look like this:
#ok  	github.com/Peripli/istio-broker-proxy	(cached)	coverage: 0.0% of statements [no tests to run]
#ok  	github.com/Peripli/istio-broker-proxy/integration	(cached)	coverage: 0.0% of statements

coverage_map = {}

test_lines = test_output.split("\n")
for test_line in test_lines:
    if len(test_line) > 0:
        test_result = re.split(r'\s+', test_line)
        coverage = test_result[-3]
        if "%" not in coverage:
            if len(test_result) >= 7:
                coverage = test_result[-7]
            else:
                coverage = "0%"
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
        print("Coverage got better. Writing new reference.")
        write_coverage(coverage_map, filename)
                
        if args.push_better_ref:
                os.chdir(src_path)
                subprocess.call(["git", "commit", "-a", "-m", "Update coverage reference."])
                subprocess.call(["git", "push", "origin", "HEAD:master"])

if not coverage_worse and not coverage_better:
        print("Coverage stayed the same.")
