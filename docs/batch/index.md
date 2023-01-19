---
id: index
title: temporal batch
sidebar_label: batch
description: nextLine
tags:
	- cli
---


Operations performed on Batch jobs. Use Workflow commands with --query flag to start batch jobs.

    Batch Jobs run in the background and affect [Workflow Executions](https://docs.temporal.io/workflows/#workflow-execution) one at a time.
    
    In `cli`, the Batch Commands are used to view the status of Batch jobs, and to terminate them.
    A successfully started Batch job returns a Job Id, which is needed to execute Batch Commands.
    
    Terminating a Batch Job does not roll back the operations already performed by the job itself.

