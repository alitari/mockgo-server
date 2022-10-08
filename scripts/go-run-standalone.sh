#!/bin/sh

cd mockgo-standalone/cmd
MOCK_DIR="../../test/main" RESPONSE_DIR="../../test/main/responses" go run . 
cd -
