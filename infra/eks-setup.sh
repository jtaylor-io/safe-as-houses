
#!/usr/bin/env bash

EKS_CLUSTER_ROLE_NAME=safe-as-houses-eks-cluster-role


aws iam create-role \
  --role-name ${EKS_CLUSTER_ROLE_NAME} \
  --assume-role-policy-document file://"eks-cluster-trust-policy.json"

aws iam attach-role-policy \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSClusterPolicy \
  --role-name ${EKS_CLUSTER_ROLE_NAME}
