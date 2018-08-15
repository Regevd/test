#!/bin/bash
export AWS_PROFILE=arch-${OWNER:-$USER}
export KOPS_STATE_STORE_BUCKET=dev-okro-io
export KOPS_STATE_STORE=s3://$KOPS_STATE_STORE_BUCKET
export NAME=k8s.${OWNER:-$USER}.dev.okro.io