#!/bin/sh
for i in taskInfo destroy sampleApply hostState; do
    go install capi_tools/${i}
done
