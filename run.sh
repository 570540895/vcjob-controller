#!/usr/bin/env bash
nohup ./vcjob-controller -kubeconfig=$HOME/.kube/config >/dev/null 2>&1 &
