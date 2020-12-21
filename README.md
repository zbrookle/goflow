
# GoFlow

[![CI](https://github.com/zbrookle/goflow/workflows/CI/badge.svg?branch=master)](https://github.com/zbrookle/goflow/actions)
[![go report](https://goreportcard.com/badge/github.com/zbrookle/goflow)](https://goreportcard.com/report/github.com/zbrookle/goflow)

GoFlow is a job scheduler inspired by Apache Airflow and is designed to be a more cohesive package,
directly integrated with Kubernetes. GoFlow parses DAGS into Kubernetes CronJobs, to remove the complexity
of scheduling from it's internal structure. Rather than run the jobs, GoFlow monitors them and ensures that
all jobs run smoothly.

In GoFlow, DAGS are supported in the following formats:

- Tradiitonal Airflow DAG (Coming soon...)
- Golang DAG (Coming soon...)

## Why GoFlow

GoFlow has the potential to surpass Apache Airflow in terms of design and scalability for a few key reasons:

1. GoFlow is written in Golang, a language that is compiled and has native support for multithreading. This
removes the dependency that Airflow has traditionally had on Redis, for scalability and drastically removes
the CPU footprint of running separate Python processes.

1. GoFlow is designed to be cohesive and less flexible than Airflow. Airflow has grown in complexity over the
years and suits a great number of use cases. However, it is very difficult to set up because of this flexibility.
GoFlow only will run on Kubernetes, meaning that it will require less configuration and gives developers back the
time they need to focus on code rather than infrastructure.

1. Because Golang is statically typed, it is much less prone to typing errors than Apache Airflow's language Python
which is dynamically typed.

1. GoFlow will have built in support for ElasticSearch logging. (Coming soon...)

## Design

GoFlow relies on a central orchestrator to maintain and monitor the CronJobs that it has placed into the Kubernetes
environment. These jobs are declared using DAGs as is traditional in Apache Airflow. GoFlow keeps track of the days
that jobs have run and can alert users if jobs fail or have not been run yet. GoFlow also collects performance
metrics from these CronJobs.

GoFlow also features a comprehensive UI, that allows users to easily create new jobs, view performance of running
jobs, and track the health of the server itself

### Job Information

GoFlow collects all DAG and DAG run information in a database for convenience and backup purposes. This information may
be useful for tracking tasks externally.

#### DAGs

Table includes:

- id
- name
- namespace
- version
- file_path
- file_format
- created_date
- last_updated_date

#### DAGRuns

Table includes:

- dag_id
- status
- execution_date
- start_date
- end_date
- last_updated_date
