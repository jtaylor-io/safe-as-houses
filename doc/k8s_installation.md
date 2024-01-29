# SCRATCH NOTES

Will need to be translated into automated setup using Terraform, etc

# Kubernetes Setup

## point at correct cluster

aws eks update-kubeconfig --name safe-as-houses --region eu-west-2

## add github role to aws-auth config map

kubectl apply -f eks/aws-auth.yaml

## ingress service routing

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/aws/deploy.yaml

## tls

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml

## remove any old load balancers (in console)

## add A record in route 53 hosted zone to point to new load balancer

## !!!!! WARNING: scratchpad cmds below this line !!!!!

## adding oidc provider to cluster

eksctl utils associate-iam-oidc-provider --cluster safe-as-houses --approve

cluster_name=safe-as-houses
oidc_id=$(aws eks describe-cluster --name $cluster_name --query "cluster.identity.oidc.issuer" --output text | cut -d '/' -f 5)
aws iam list-open-id-connect-providers | grep $oidc_id | cut -d "/" -f4
eksctl utils associate-iam-oidc-provider --cluster $cluster_name --approve
aws iam list-attached-role-policies --role-name my-github-actions-role --query AttachedPolicies[].PolicyArn --output text
eksctl get iamidentitymapping --cluster $cluster_name
kubectl describe configmap aws-auth -n kube-system
kubectl apply -f eks/aws-auth.yaml
