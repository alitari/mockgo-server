#!/bin/sh

# check if hey is installed
if ! command -v hey &> /dev/null
then
    echo "hey could not be found, please install it first"
    exit
fi

# check if parameters are supplied
if [[ -z $1 || -z $2 || -z $3 || -z $4 || -z $5 || -z $6 || -z $7 ]]; then
    echo "usage: load-test.sh <count> <concurrent> <quote> <host> <method> <path> <endpoint>"
    exit 1
fi

count=$1
concurrent=$2
quote=$3
host=$4
method=$5
path=$6
endpoint=$7
if [ -t 0 ]; then
    payload=""
else
    payload=$(cat)
fi

loadtest_dir=test/load-test
password_file=$loadtest_dir/.mockgo_password

# read password variable from file ".mockgo_password"
if [ -f $password_file ]; then
    mockgo_password=$(cat $password_file)
else
    echo "no '$password_file' file found, please create one first"
    exit 1
fi

url="http://${host}${path}"

echo "Running $count requests with $concurrent workers and $quote requests per second"
echo "Request: $method $url"
echo "Payload: '$payload'"

hurl $loadtest_dir/cleanup.hurl --variable mockgo_host=$host --variable endpoint=$endpoint --variable mockgo_password=$mockgo_password --test

hey -n $count -c $concurrent -q $quote -m $method $url -d $payload

hurl $loadtest_dir/check.hurl --variable mockgo_host=$host --variable endpoint=$endpoint --variable mockgo_password=$mockgo_password --variable matches_count=$count --variable mismatches_count="0" --test

