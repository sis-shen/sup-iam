# iam-api-server
kubectl port-forward --address 0.0.0.0 service/iam-api-server 8080:8080 -n iam
# iam-auth-server
kubectl port-forward --address 0.0.0.0 service/iam-auth-server 8081:8080 -n iam

kubectl port-forward --address 0.0.0.0 service/iam-auth-server 7070:7070 -n iam

# MySQL 注入数据用
kubectl port-forward --address 0.0.0.0 service/mysql-dev-primary 33066:3306 -n iam