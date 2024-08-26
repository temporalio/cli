#!/bin/bash

for i in {1..20}
do
    echo "Run #$i"
    go test -v -run TestSharedServerSuite/TestTaskQueue_Describe_Task_Queue_Stats_NonEmpty
    if [ $? -ne 0 ]; then
        echo "Test failed on run #$i"
        exit 1
    fi
done

echo "All 100 runs completed successfully."