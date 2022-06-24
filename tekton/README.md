## Tekton

This is a living example of how to use the CLI in Tekton Pipelines.

### Prerequisites

* Kubernetes / openshift cluster
* Pipelines enabled (something which provides the `tekton.dev/v1beta1` `apiVersion`)


### Architecture

* The cron job pull down the latest DB and uncompresses it into a PVC
* The `Pipeline` will then index the image and then match it to vulns in the DB in the PVC

### Quick Start

\*note: using `oc` and `kubectl` interchangably

```bash
$ oc apply -f pvc.yaml  # create the PVC
$ oc apply -f cron.yaml  # Define cron, you might need to adjust the schedule to trigger
$ oc apply -f pipeline.yaml  # Define the pipeline
$ oc apply -f pipeline_run.yaml  # Start a pipeline run
```

### TODOs

* Parameterize inputs
