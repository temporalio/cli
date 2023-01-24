---
id: index
title: temporal batch
sidebar_label: batch
description: Operations performed on Batch jobs. Use Workflow commands with --query flag to start batch jobs.
tags:
	- cli
---


Batch commands allow you to change multiple [Workflow Executions](/concepts/what-is-a-workflow-execution) without having to repeat yourself on the command line.
In order to do this, you provide the command with a [List Filter](/concepts/what-is-a-list-filter) and the type of Batch job to execute.

The List Filter identifies the Workflow Executions that will be affected by the Batch job.
The Batch type determines the other parameters that need to be provided, along with what is being affected on the Workflow Executions.

To start the Batch job, run `temporal workflow query`.
A successfully started Batch job will return a Job ID.
Use this Job ID to execute other actions on the Batch job.

