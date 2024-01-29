# SCRATCH NOTES

Will need to be translated into automated setup using Terraform, etc

# Kubernetes Setup

## point at correct cluster

```
aws eks update-kubeconfig --name safe-as-houses --region eu-west-2
```

## add github role to aws-auth config map

```
kubectl apply -f eks/aws-auth.yaml
```

## install nginx ingress controller

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/aws/deploy.yaml
```

## install cert-manager

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml
```

## remove any old load balancers (in console)

## add A record in route 53 hosted zone to point to new load balancer
